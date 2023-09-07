package main

import (
	"context"
	"flag"
	"fmt"
	"path/filepath"
	"time"

	machineapi "github.com/openshift/api/machine/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func main() {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}

	maxAgeDays := flag.Float64("max-age-days", 28, "an int")
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}

	s := runtime.NewScheme()
	utilruntime.Must(corev1.AddToScheme(s))
	utilruntime.Must(machineapi.AddToScheme(s))

	c, err := client.New(config, client.Options{
		Scheme: s,
	})
	if err != nil {
		panic(err)
	}

	machineList := &machineapi.MachineList{}

	err = c.List(context.Background(), machineList, &client.ListOptions{Namespace: "openshift-machine-api"})

	if err != nil {
		panic(err)
	}

	today := time.Now()

	//oldestMachineIndex := 0
	oldestMachineAge := float64(0)
	oldestMachine := machineList.Items[0]

	fmt.Println("Machines that are older than", *maxAgeDays, " days, that are not in Deleting state:\n")

	for i, machine := range machineList.Items {

		created := machine.CreationTimestamp
		age := today.Sub(created.Time).Hours() / 24

		if age > *maxAgeDays && machine.DeletionTimestamp == nil {

			if age > oldestMachineAge {
				oldestMachineAge = age
				//oldestMachineIndex = i
				oldestMachine = machineList.Items[i]
			}

			fmt.Println(machine.Name)
			fmt.Println(machine.Annotations)
			fmt.Println(machine.Labels)

		}

	}

	//Found the oldest machine
	fmt.Println("\n Found the oldest machine", oldestMachine.Name, "aged", oldestMachineAge)

	if oldestMachine.Labels["machine.openshift.io/cluster-api-machine-role"] == "worker" {

		fmt.Println("\n The oldest machine", oldestMachine.Name, "is a worker node. Would you like to proceed to delte? Y/N:")

	}

	if oldestMachine.Labels["machine.openshift.io/cluster-api-machine-role"] == "master" {

		fmt.Println("\n The oldest machine", oldestMachine.Name, "is a master node. Would you like to proceed to delte? Y/N:")

	}

}
