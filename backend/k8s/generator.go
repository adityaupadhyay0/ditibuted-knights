package k8s

import (
	"fmt"
	"strings"

	"github.com/user/distributed-knights/backend/parser"
)

func GenerateManifests(config *parser.InfraConfig) string {
	var manifests []string

	for _, res := range config.Resources {
		switch res.Type {
		case "aws_instance":
			manifests = append(manifests, generateDeployment(res.Name))
		case "aws_db_instance":
			manifests = append(manifests, generateStatefulSet(res.Name))
		case "aws_vpc":
			manifests = append(manifests, generateNamespace(res.Name))
		}
	}

	return strings.Join(manifests, "\n---\n")
}

func generateDeployment(name string) string {
	return fmt.Sprintf(`apiVersion: apps/v1
kind: Deployment
metadata:
  name: %s
spec:
  replicas: 1
  selector:
    matchLabels:
      app: %s
  template:
    metadata:
      labels:
        app: %s
    spec:
      containers:
      - name: main
        image: nginx:latest`, name, name, name)
}

func generateStatefulSet(name string) string {
	return fmt.Sprintf(`apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: %s
spec:
  serviceName: %s
  replicas: 1
  selector:
    matchLabels:
      app: %s
  template:
    metadata:
      labels:
        app: %s
    spec:
      containers:
      - name: db
        image: postgres:latest`, name, name, name, name)
}

func generateNamespace(name string) string {
	return fmt.Sprintf(`apiVersion: v1
kind: Namespace
metadata:
  name: %s`, name)
}
