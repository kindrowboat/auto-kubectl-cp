package main

import (
	"flag"
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
)

// syncFileToPods copies the modified file to all pods in the specified deployment
func syncFileToPods(filePath, deploymentName, containerName, containerPath, namespace string) {
	// Get pod names in the deployment
	pods, err := getPodsInDeployment(deploymentName, namespace)
	if err != nil {
		log.Fatalf("Error getting pods: %v", err)
	}

	// Copy file to each pod
	for _, pod := range pods {
		err := copyFileToContainer(pod, filePath, containerName, containerPath, namespace)
		if err != nil {
			log.Printf("Error copying file to pod %s: %v", pod, err)
		} else {
			log.Printf("Copied %s to %s:%s", filePath, pod, containerPath)
		}
	}
}

// getPodsInDeployment retrieves the pod names in the specified deployment
func getPodsInDeployment(deploymentName, namespace string) ([]string, error) {
	var kubectlArgs []string
	kubectlArgs = append(kubectlArgs, "get", "pods", "-l", fmt.Sprintf("app=%s", deploymentName), "-o", "jsonpath={.items[*].metadata.name}")
	if namespace != "" {
		kubectlArgs = append(kubectlArgs, "-n", namespace)
	}

	cmd := exec.Command("kubectl", kubectlArgs...)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	pods := strings.Fields(string(output))
	return pods, nil
}

// copyFileToContainer copies a file to the specified container in a pod
func copyFileToContainer(pod, filePath, containerName, containerPath, namespace string) error {
	fileName := filepath.Base(filePath)
	var kubectlArgs []string
	kubectlArgs = append(kubectlArgs, "cp", filePath, fmt.Sprintf("%s:%s/%s", pod, containerPath, fileName), "-c", containerName)
	if namespace != "" {
		kubectlArgs = append(kubectlArgs, "-n", namespace)
	}
	cmd := exec.Command("kubectl", kubectlArgs...)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("kubectl cp failed: %v", err)
	}
	return nil
}

// monitorDirectory sets up a file watcher on the local directory and syncs files on modification or creation
func monitorDirectory(localPath, deploymentName, containerName, containerPath, namespace string) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("error creating watcher: %v", err)
	}
	defer watcher.Close()

	done := make(chan bool)

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
					log.Printf("File changed: %s", event.Name)
					syncFileToPods(event.Name, deploymentName, containerName, containerPath, namespace)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Printf("Error: %v", err)
			}
		}
	}()

	err = watcher.Add(localPath)
	if err != nil {
		return fmt.Errorf("error watching directory: %v", err)
	}

	log.Printf("Monitoring %s for changes...", localPath)
	<-done
	return nil
}

func main() {
	// Parse command-line arguments with both short and long flags
	localPath := flag.String("local-path", "", "Local path to monitor for file changes")
	deploymentName := flag.String("deployment", "", "Kubernetes deployment name")
	containerName := flag.String("container", "", "Container name in the deployment's pods")
	containerPath := flag.String("container-path", "", "Path inside the container to copy files to")
	namespace := flag.String("namespace", "", "Kubernetes namespace")

	flag.Parse()

	if *localPath == "" || *deploymentName == "" || *containerName == "" || *containerPath == "" {
		log.Fatal("All arguments except namespace are required. Use -h for help.")
	}

	// Monitor the directory and sync files
	err := monitorDirectory(*localPath, *deploymentName, *containerName, *containerPath, *namespace)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
}
