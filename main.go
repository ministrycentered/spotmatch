package main

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
)

func main() {
	lambda.Start(Handler)
}

type Event struct {
	SpotASGName     string `json:"SpotASGName"`
	OnDemandASGName string `json:"OnDemandASGName"`
}

func Handler(ctx context.Context, event Event) {
	spotASG := fetchASG(event.SpotASGName)

	// Consider InService and Pending in the count, as they are active, running
	// EC2 instances. Even if not yet done with all lifecycle events, they will
	// be soon and we don't want to add more capacity.
	var spotASGInServiceInstances []autoscaling.Instance
	for _, instance := range spotASG.Instances {
		if *instance.LifecycleState == "InService" || *instance.LifecycleState == "Pending" {
			spotASGInServiceInstances = append(spotASGInServiceInstances, *instance)
		}
	}

	// If the Spot group is good to go, set the on-demand desired capacity to 0
	// and exit.
	if int64(len(spotASGInServiceInstances)) >= *spotASG.DesiredCapacity {
		fmt.Println(fmt.Sprintf("spot ASG '%s' has sufficient capacity, setting on-demand ASG '%s' to 0 capacity", event.SpotASGName, event.OnDemandASGName))
		setDesiredCapacity(event.OnDemandASGName, 0)
		return
	}

	// Make up for any missing capacity in Spot with On-Demand
	supplementalOnDemandCapacity := *spotASG.DesiredCapacity - int64(len(spotASGInServiceInstances))
	fmt.Println(fmt.Sprintf("spot ASG '%s' has insufficient capacity, setting on-demand ASG '%s' to %v capacity", event.SpotASGName, event.OnDemandASGName, supplementalOnDemandCapacity))
	setDesiredCapacity(event.OnDemandASGName, supplementalOnDemandCapacity)
}

func fetchASG(asgName string) autoscaling.Group {
	session, _ := session.NewSession(&aws.Config{Region: aws.String("us-east-1")})
	autoscalingSession := autoscaling.New(session)

	input := &autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []*string{aws.String(asgName)},
	}
	output, err := autoscalingSession.DescribeAutoScalingGroups(input)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// This should only happen if you misspell the ASG name or it gets deleted
	// before the Lambda function is deleted.
	if len(output.AutoScalingGroups) == 0 {
		fmt.Println(fmt.Sprintf("Requested ASG '%s' could not be found", asgName))
		os.Exit(1)
	}

	return *output.AutoScalingGroups[0]
}

func setDesiredCapacity(asgName string, capacity int64) {
	session, _ := session.NewSession(&aws.Config{Region: aws.String("us-east-1")})
	autoscalingSession := autoscaling.New(session)

	input := &autoscaling.UpdateAutoScalingGroupInput{
		AutoScalingGroupName: aws.String(asgName),
		DesiredCapacity:      aws.Int64(capacity),
	}

	_, err := autoscalingSession.UpdateAutoScalingGroup(input)

	if err != nil {
		fmt.Println(err)
	}
}
