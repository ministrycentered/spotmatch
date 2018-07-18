# Spotmatch

Lambda function for ensuring a Spot ASG maintains desired capacity by falling back to capacity in On-Demand group

## Run locally with SAM (requires real ASGs specified in event.json)

```
sam local invoke --event event.json --profile myawsprofile
```

## Package with SAM

```
sam package --template-file template.yml --s3-bucket mysamartifactss3bucket --output-template-file packaged.yml --profile myawsprofile
```

The output of `sam package` includes a CloudFormation template. Use this template in a new CloudFormation stack and you're good to go!

## Deploy

[![Launch stack in us-east-1](https://s3.amazonaws.com/cloudformation-examples/cloudformation-launch-stack.png)](https://console.aws.amazon.com/cloudformation/home?region=us-east-1#/stacks/new?stackName=spotmatch&templateURL=https://s3.amazonaws.com/cf-spotmatch/v1.0/packaged.yml)

## Usage

The CloudFormation stack exports a global `SpotMatchFunction` that you can import with `ImportValue` into other stacks. From another CloudFormation template, setup a recurring task that runs every minute to verify capacity is met.

`OnDemandServerGroup` and `SpotServerGroup` are references to ASGs elsewhere in the template.

```yml
SpotMatchRecurringTask:
  Type: AWS::Events::Rule
  Properties:
    Description: Ensure ASGs contain desired capacity
    ScheduleExpression: rate(1 minute)
    State: ENABLED
    Targets:
        - Arn: !Join ['', ['arn:aws:lambda:us-east-1:569307328219:function:', !ImportValue SpotMatchFunction, ':live']]
          Id: spotmatch-recurring-task
          Input: !Sub '{"SpotASGName": "${SpotServerGroup}", "OnDemandASGName":"${OnDemandServerGroup}"}'
PermissionForEventsToInvokeSpotMatchLambdaFunction:
  Type: AWS::Lambda::Permission
  Properties:
    FunctionName: !Join ['', ['arn:aws:lambda:us-east-1:569307328219:function:', !ImportValue SpotMatchFunction, ':live']]
    Action: lambda:InvokeFunction
    Principal: events.amazonaws.com
    SourceArn: !GetAtt SpotMatchRecurringTask.Arn
```
