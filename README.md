# jx-app-sonar-scanner

jx-app-sonar-scanner invokes the SonarQube scanner client following the make-build step of every pipeline.

## Installation

Using the [jx command line tool](https://jenkins-x.io/getting-started/install/), run the following command:

```bash
$ jx add app jx-app-sonar-scanner
```

After the installation, you can view the status of jx-app-sonar-scanner via:

```bash
$ jx get app jx-app-sonar-scanner
```

You will be asked for:

- The fully qualified address of the Sonarqube instance which is typically something like 'http://jx-sonarqube.sonarqube.svc.cluster.local:9000'
- Your Sonarqube user token
- Whether you would like to enable or disable scanning for preview and release builds.

## Uninstall
You can uninstall using `jx delete app jx-app-sonar-scanner`

## Additional Configuration
You will need a configured instance of a Sonarqube server.

You can provide Sonarqube related properties to the scanner by including an appropriately configured `sonar-project.properties` file in the root of your project folder.

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

Where `verbose` turns on logging within the pipeline. `skip` causes scanning to be skipped for this project. `pullRequest` and `release` specify the pipeline, stage and step AFTER which you wish to insert the scan operation.