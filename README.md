# jx-app-sonar-scanner

jx-app-sonar-scanner enables you to inject the SonarQube scanner client into every pipeline that runs on your Jenkins-X instance, allowing you to set up automated governance processes for code quality checking across your teams. It operates by dynamically re-writing your effective pipeline during the metapipeline phase of the build, creating a new step within the desired stage of your build that invokes the SonarQube scanner client in its own container.

Out-of-the-box, jx-app-sonar-scanner does its best to recognise the build environment you are using for each project and insert the scanner step into the most appropriate place in your pipelines. If you are using a custom build pipeline config, you can also tell jx-app-sonar-scanner where it should execute in your pipeline.

You will get basic scanning capabilities automatically, however many of the default pipelines in the current build packs do not include linting, unit testing or code coverage actions so you will want to extend your build pipelines to include these to use the full capabilities of SonarQube.

## Installation
Prerequisites: You will need a configured instance of a SonarQube server. jx-app-sonar-scanner will work with the community edition of SonarQube. If you do not have an instance already set up, you can use this [SonarQube](https://github.com/Oteemo/charts/tree/master/charts/sonarqube) chart to help get you going. If installing on your Kubernetes cluster with Jenkins-X, it is suggested that you create a dedicated `sonarqube` environment and use GitOps to manage the installation. SonarQube is very much a traditional, stateful, client-server application that does not currently sit easily within a containerized environment, so some care will be needed if operating in this configuration.

To set up jx-app-sonar-scanner:

Using the [jx command line tool](https://jenkins-x.io/getting-started/install/), run the following command:

```bash
$ jx add app jx-app-sonar-scanner
```

After the installation, you can view the status of jx-app-sonar-scanner via:

```bash
$ jx get app jx-app-sonar-scanner
```

You will be asked for:

- The fully qualified address of the SonarQube instance which is typically something like 'http://jx-sonarqube.sonarqube.svc.cluster.local:9000'
- Your Sonarqube user token
- Whether you would like to enable or disable scanning for preview or release builds.

## Uninstall
You can uninstall using `jx delete app jx-app-sonar-scanner`

## Custom Configuration
You can provide SonarQube related properties to the scanner by including an appropriately configured `sonar-project.properties` file in the root of your project folder. This will be necessary if you want to customise the SonarQube plugins you are using on your server instance.

In addition, you can alter the configuration of the scanner within the Jenkins-X pipeline by providing a `.jx-app-sonar-scanner.yaml` file in the root of your project folder.

This should be of the form:

```yaml
---
verbose: true
skip: false
pullRequest:
    stage: build
    step: build-container-build
release:
    stage: build
    step: build-container-create
```

Where `verbose` turns on logging within the pipeline. `skip` causes scanning to be skipped for this project. `pullRequest` and `release` specify the pipeline, stage and step AFTER which you wish to insert the scan operation. You can use this feature to support custom pipeline configs or build packs that are not recognised by default. If you are the creator of a public build pack, please feel free to submit a PR to add detection for your pack to the app.

All top-level terms are optional.

`skip` creates an entry in the build log, declaring that quality checking has been skipped for a given project, so it remains possible to detect exceptions to your governance processes.