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

