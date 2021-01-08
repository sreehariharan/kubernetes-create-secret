package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	corev1Types "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	secretsClient corev1Types.SecretInterface
)

func createOrUpdateSecret(clientset *kubernetes.Clientset, secretSpec *corev1.Secret) {

	var operation string
	var err error

	if secretSpec.Namespace == "" {
		secretSpec.Namespace = "default"
	}

	secretsClient = clientset.CoreV1().Secrets(secretSpec.Namespace)

	existSec, _ := secretsClient.Get(context.TODO(), secretSpec.Name, metav1.GetOptions{})

	if len(existSec.Name) == 0 {
		_, err = secretsClient.Create(context.TODO(), secretSpec, metav1.CreateOptions{})
		operation = "created"
	} else {
		_, err = secretsClient.Update(context.TODO(), secretSpec, metav1.UpdateOptions{})
		operation = "updated"
	}

	if err != nil {
		panic(err.Error())
	} else {
		fmt.Printf("secret %s %s successfully", secretSpec.Name, operation)
	}

}

func parseSecretDataJSONFile(jsonFilePath *string) map[string]string {

	returnMap := make(map[string]string)

	if *jsonFilePath != "" {

		jsonFile, err := os.Open(*jsonFilePath)
		if err != nil {
			panic(err.Error())
		}
		// defer the closing of our jsonFile so that we can parse it later on
		defer jsonFile.Close()

		byteValue, _ := ioutil.ReadAll(jsonFile)

		json.Unmarshal([]byte(byteValue), &returnMap)

	}
	return returnMap
}

func main() {

	var (
		kubeconfig         = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
		secretName         = flag.String("name", "default-sec", "name of the secret to create")
		secretNamespace    = flag.String("namespace", "default", "namespace of the secret")
		secretType         = flag.String("type", string(corev1.SecretTypeOpaque), "type of the secret")
		secretDataJSONPath = flag.String("data-json-path", "", "absolute path to secret data in json format")

		err    error
		config *rest.Config
	)

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "usage %s", os.Args[0])
		fmt.Println("")
		flag.PrintDefaults()
	}

	flag.Parse()

	secretData := parseSecretDataJSONFile(secretDataJSONPath)

	if *kubeconfig == "" {
		fmt.Println("Unable to locate kubeconfig, trying to use In Cluster connection")
		config, err = rest.InClusterConfig()
		if err != nil {
			panic(err.Error())
		}
	} else {
		// use the current context in kubeconfig
		config, err = clientcmd.BuildConfigFromFlags("", *kubeconfig)
		if err != nil {
			panic(err.Error())
		}
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	secretSpec := corev1.Secret{}
	secretSpec.Type = corev1.SecretType(*secretType)
	secretSpec.Name = *secretName
	secretSpec.Namespace = *secretNamespace
	secretSpec.StringData = secretData

	createOrUpdateSecret(clientset, &secretSpec)

}
