package main

import (
	"fmt"
	"flag"
	"time"
	"regexp"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"k8s.io/client-go/rest"
)

var builddate string

func main() {
	// creates the in-cluster config
	fmt.Printf("Trying InClusterConfig first\n")
	config, err := rest.InClusterConfig()
	if err != nil {
		fmt.Printf("InClusterConfig failed with error %+v, trying --kubeconfig option next\n", err)

		kubeconfig := flag.String("kubeconfig", "./config", "absolute path to the kubeconfig file")
		flag.Parse()
		// uses the current context in kubeconfig
		config, err = clientcmd.BuildConfigFromFlags("", *kubeconfig)
		if err != nil {
			panic(err.Error())
		}
		
	}
	
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	// capture region (not availability zone) and instance id from "aws:///us-east-1d/i-00cd1777ad2f664f0""
	instance_regex, err := regexp.Compile("^aws:///(.+?)[a-z]/(i-.+)$")
	if err != nil {
		panic(err)
	}
	for {
		nodes, err := clientset.CoreV1().Nodes().List(metav1.ListOptions{})
		if err != nil {
			panic(err.Error())
		}
		fmt.Printf("There are %d nodes in the cluster\n", len(nodes.Items))

		for _, node := range nodes.Items {
			fmt.Printf("checking node %s\n", node.ObjectMeta.Name)
			value, ok := node.ObjectMeta.Labels["unity3d.com/sourcedestcheckdisabled"]
			if ok && value == "true" {
				fmt.Printf("sourcedestcheck already disabled from machine %s\n", node.ObjectMeta.Name)
				continue
			}

			m := instance_regex.FindStringSubmatch(node.Spec.ProviderID)
			if len(m) > 0 {
				fmt.Printf("Disabling source/dest check: %s instance id is %s\n", m[1], m[2])

				disabled, err := disableSourceDestCheck(m[1], m[2])
				if err == nil && disabled == true {
					fmt.Printf("Disabled. Adding label unity3d.com/sourcedestcheckdisabled=true to %s\n", node.ObjectMeta.Name)
					node.ObjectMeta.Labels["unity3d.com/sourcedestcheckdisabled"] = "true"
					_, err := clientset.CoreV1().Nodes().Update(&node)
					if err != nil {
						fmt.Printf("Error adding label to %s. This should not be fatal unless this happens constantly. Error: %+v\n", node.ObjectMeta.Name, err)
					}
				}
			}
		}

		time.Sleep(10 * time.Second)
	}
}

func disableSourceDestCheck(awsRegion string, instance_id string) (bool, error) {
	sess := session.Must(session.NewSession())

	svc := ec2.New(sess, &aws.Config{Region: aws.String(awsRegion)})
	params := &ec2.DescribeInstancesInput{
		InstanceIds: []*string{&instance_id},
	}
	resp, err := svc.DescribeInstances(params)
	if err != nil {
		fmt.Println("there was an error listing instances in", awsRegion, err.Error())
		panic(err)
	}

	instance := *resp.Reservations[0].Instances[0]
	eni := instance.NetworkInterfaces[0]
	if *eni.SourceDestCheck == true {
		fmt.Printf("Turning SourceDestCheck off on network interface %s (instance %s)", *eni.NetworkInterfaceId, instance_id)

		params := &ec2.ModifyNetworkInterfaceAttributeInput{
			NetworkInterfaceId: eni.NetworkInterfaceId, // Required
			SourceDestCheck: &ec2.AttributeBooleanValue{
				Value: aws.Bool(false),
			},
		}

		resp2, err := svc.ModifyNetworkInterfaceAttribute(params)
		if err != nil {
			fmt.Printf("Error disabling source/dest check on %s\n", eni.NetworkInterfaceId)
			panic(err)
		}
		fmt.Printf("ModifyNetworkInterfaceAttribute: %+v\n", *resp2)
		return true, nil
	}

	if *eni.SourceDestCheck == false {
		return true, nil
	}

	return false, nil
}


