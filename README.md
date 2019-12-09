# `ctsns`

`ctsns` is a [Certificate Transparency][ct] (CT) to [AWS SNS][sns] gateway. 
Whenever a new TLS certificate is issued and logged to the CT logs, a message
will be published to a publicly subscribable SNS topic. There is one message
per certificate. 

Additionally, the list of domains in the certificate is added to the SNS message 
as message attributes, allowing subscribers to [filter][filter] received messages 
to only the domains they are interested in.

Existing solutions already exist in some form or another, like Facebook's webhooks,
etc. but I wanted something that I could trivially integrate into my existing AWS
ecosystem. It made sense to share publicly.

## How do I use it?

The SNS topic ARN is `TODO`. You can subscribe either Lambda functions or SQS queues
to it. Webhooks, SMS and email are unsupported as they cost me extra money and I'm
doing this out of my own pocket. 

## Self-managed

If you *need* those additional SNS subscriber types, this project is open source
and can be run out of your own account with minimal setup. It runs as a Fargate Spot
task and costs less than $3/month for the task and about $45/month for *publishing*
to the SNS topic.

## How does it work?

Major kudos to Calidog for both publishing and hosting [Certstream][certstream].
Certstream does 99% of the heavy lifting here, in that it polls the public CT logs
and pushes certificate events to a websocket in simple JSON format. 

[ct]: https://www.certificate-transparency.org/ 
[sns]: https://aws.amazon.com/sns/
[filter]: https://docs.aws.amazon.com/sns/latest/dg/sns-message-filtering.html
[certstream]: https://certstream.calidog.io/
