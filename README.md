# JX 

JX is a command line tool for installing and using [Jenkins X](https://jenkins-x.io/).

[![CII Best Practices](https://bestpractices.coreinfrastructure.org/projects/3237/badge)](https://bestpractices.coreinfrastructure.org/projects/3237)
[![GoDoc](https://godoc.org/github.com/jenkins-x/jx?status.svg)](https://godoc.org/github.com/jenkins-x/jx)
[![Docs](https://readthedocs.org/projects/docs/badge/?version=latest)](https://jenkins-x.io/documentation/)
[![Docker Pulls](https://img.shields.io/docker/pulls/jenkinsxio/jx.svg)](https://hub.docker.com/r/jenkinsxio/jx/tags)
[![Downloads](https://img.shields.io/github/downloads/jenkins-x/jx/total.svg)](https://github.com/jenkins-x/jx/releases)
[![GoReport](https://goreportcard.com/badge/github.com/jenkins-x/jx)](https://goreportcard.com/report/github.com/jenkins-x/jx)
[![Apache](https://img.shields.io/badge/license-Apache-blue.svg)](https://github.com/jenkins-x/jx/blob/master/LICENSE)
[![Reviewed by Hound](https://img.shields.io/badge/reviewed_by-Hound-8E64B0.svg)](https://houndci.com)
![Build Status](https://img.shields.io/endpoint?url=https%3A%2F%2Fstatusbadge-jx.jenkins-x.live%2Fjx)

## Installing

Check out [how to install jx](https://jenkins-x.io/docs/getting-started/setup/install/).

## Getting Help

To find out the available commands type:

    jx

Or to get help on a specific command, say, `create` then type:

    jx help create

You can also browse the [jx command reference documentation](https://jenkins-x.io/commands/jx/).

## Getting Started

Please check out the [Getting Started Guide](https://jenkins-x.io/docs/getting-started/) on how to:

* [create new Kubernetes clusters with Jenkins X](https://jenkins-x.io/docs/getting-started/setup/create-cluster/)
* [install Jenkins X on existing Kubernetes clusters](https://jenkins-x.io/docs/managing-jx/old/install-on-cluster/)

Then [what to do next when you have Jenkins X installed](https://jenkins-x.io/getting-started/next/).
    
## Reference

* [Command Line Reference](https://jenkins-x.io/commands/jx/#jx)
  
  
## Opening Consoles

To open the Jenkins console try:

    jx console
    
Or to open other consoles:

    jx open foo
    
If you do not know the name:

    jx open
    

## Tail logs

To tail the logs of anything running on Kubernetes (jenkins or your own applications) type.

    jx logs
    
Which prompts you for the deployment to log then tails the logs of the newest pod for an app.

You can filter the list of deployments via:

    jx logs -f cheese

Then if there's only one deployment with a name that contains `cheese` then it'll tail the logs of the latest pod or will prompt you to choose the exact deployment to use.

## Remote shells

You can open a remote shell inside any pods container via the `rsh` command:

    jx rsh
    
Or to open a shell inside a pod named foo:

    jx rsh foo

Pass `-c` to specify the container name. e.g. to open a shell in a maven build pod:

    jx rsh -c maven maven

## Importing or Creating apps

To import an application from the current directory:

    jx import
    
Or to create a new Spring Boot application from scratch:

    jx create spring
    
e.g. to create a new WebMVC and Spring Boot Actuator microservice try this:

    jx create spring -d web -d actuator
        
If you have a Maven Archetype you would like to create then use:

    jx create archetype
    
## Starting builds

To start a pipeline using a specific name try:

    jx start pipeline myorg/myrepo

Or to pick the pipeline to start:

    jx start pipeline

If you know part of the name of the pipeline to run you can filter the list via:

    jx start pipeline -f thingy

You can start and tail the build log via:
		
    jx start pipeline -t

## Viewing Apps and Environments

To view environments for a team:

    jx get env
    
To view the application versions across environments:

    jx get version
            
## Manual promotions

Typically we setup Environments to use _automatic_ promotion so that the CI / CD pipelines will automatically promote versions through the available Environments using the CI / CD Pipeline.

However if you wish to manually promote a version to an environment you can use the following command:

    jx promote myapp -e prod 
    
Or if you wish to use a custom namespace:   

    jx promote myapp -n my-dummy-namespace
 
## Switching Environments

The `jx` CLI tool uses the same Kubernetes cluster and namespace context as `kubectl`. 

You can switch Environments via:

    jx env
    
Or change it via:

    jx env staging
    jx env prod
    
To display the current environment without trying to change it:

    jx env -b

To view all the environments type:

    jx get env
    
You can create or edit environments too:

    jx create env
    
    jx edit env staging
    
You can switch namespaces in the same way via:

    jx ns

or

    jx ns awesome-staging    

## Switching Clusters

If you have multiple Kubernetes clusters (e.g. you are using GKE and Minikube together) then you can switch between them via:

    jx ctx
    
In the same way. Or via

    jx ctx minikube
            
**Note** that changing the namespace ,environment or cluster changes the current context for **ALL** shells!

### Sub shells

So if you want to work temporarily with, say, the production cluster we highly recommend you use a sub shell for that.

    jx shell my-prod-context
    
Or to pick the context to use for the sub shell:

    jx shell 

Then your bash prompt will be updated to reflect that you are in a different context and/or namespace. Any changes to the namespace, environment or context will be local to the current shell only!    

### Setting your prompt

You can use the `jx prompt` to configure your CLI prompt to display the current team and environment you are working within:          
                                            
		# Enable the prompt for bash
		PS1="[\u@\h \W \$(jx prompt)]\$ "

		# Enable the prompt for zsh
		PROMPT='$(jx prompt)'$PROMPT

**Note** that the prompt is updated automatically for you via the `jx shell` command too.

### Bash completion

On a Mac to enable bash completion try:

    jx completion bash > ~/.jx/bash
    source ~/.jx/bash   
    
Or try:

    source <(jx completion bash)

For more help try:

    jx help completion bash
           
## Addons

We are adding a number of addon capabilities to Jenkins X. To add or remove addons use the `jx create addon` or `jx delete addon` commands.

For example to add the [Gitea Git server](https://gitea.io/en-US/) to your Jenkins X installation try:

    jx create addon gitea

This will: 

* install the Gitea Helm chart.
* add Gitea as a Git server (via the `jx create git server gitea` command).
* create a new user in Gitea (via the `jx create git user -n gitea` command).
* create a new Git API token in Gitea (via the `jx create git token -n gitea -p password username` command).
     
## Troubleshooting

We have tried to collate common issues here with work arounds. If your issue isn't listed here please [let us know](https://github.com/jenkins-x/jx/issues/new).
 
### Cannot create cluster Minikube
If you are using a Mac then `hyperkit` is the best VM driver to use - but does require you to install a recent [Docker for Mac](https://docs.docker.com/docker-for-mac/install/) first. Maybe try that then retry `jx create cluster minikube`?

If your Minikube is failing to startup then you could try:
 
    minikube delete
    rm -rf ~/.minikube

If the `rm` fails you may need to do:

    sudo rm -rf ~/.minikube

Now try `jx create cluster minikube` again - did that help? Sometimes there are stale certs or files hanging around from old installations of minikube that can break things.

Sometimes a reboot can help in cases where virtualisation goes wrong ;)

Otherwise you could try follow the Minikube instructions:

* [install Minikube](https://github.com/kubernetes/minikube#installation)
* [run Minikube start](https://github.com/kubernetes/minikube#quickstart) 

### Minikube and hyperkit: Could not find an IP address

If you are using Minikube on a mac with hyperkit and find Minikube fails to start with a log like:

```
Temporary Error: Could not find an IP address for 46:0:41:86:41:6e
Temporary Error: Could not find an IP address for 46:0:41:86:41:6e
Temporary Error: Could not find an IP address for 46:0:41:86:41:6e
Temporary Error: Could not find an IP address for 46:0:41:86:41:6e
```

It could be you have hit [this issue in Minikube and hyperkit](https://github.com/kubernetes/minikube/issues/1926#issuecomment-356378525).

The work around is to try the following:

```
rm ~/.minikube/machines/minikube/hyperkit.pid
``` 

Then try again. Hopefully this time it will work!

### Cannot access services on Minikube

When running Minikube locally `jx` defaults to using [nip.io](http://nip.io/) as a way of using nice-isn DNS names for services and working around the fact that most laptops can't do wildcard DNS. However sometimes [nip.io](http://nip.io/) has issues and does not work.

To avoid using [nip.io](http://nip.io/) you can do the following:

Edit the file `~/.jx/cloud-environments/env-minikube/myvalues.yaml` and add the following content:

```yaml
expose:
  Args:
    - --exposer
    - NodePort
    - --http
    - "true"
```

Then re-run `jx install` and this will switch the services to be exposed on `node ports` instead of using ingress and DNS.

So if you type:

```
jx open
```

You'll see all the URLs of the form `http://$(minikube ip):somePortNumber` which then avoids going through [nip.io](http://nip.io/) - it just means the URLs are a little more cryptic using magic port numbers rather than simple host names.


 
### Other issues

Please [let us know](https://github.com/jenkins-x/jx/issues/new) and see if we can help? Good luck! 
	
## Contributing

If you're looking to build from source or get started hacking on jx, please see the
[hacking guide][hacking] for more information.


[![](https://codescene.io/projects/4772/status.svg) Get more details at **codescene.io**.](https://codescene.io/projects/4772/jobs/latest-successful/results)



[hacking]: docs/contributing/hacking.md
