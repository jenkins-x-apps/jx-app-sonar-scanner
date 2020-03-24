while getopts s:k:r:p: option
do
case "${option}"
in
s) export SONARQUBE_SERVER=${OPTARG};;
k) export SONAR_TOKEN=${OPTARG};;
r) export SCAN_ON_RELEASE=${OPTARG};;
p) export SCAN_ON_PREVIEW=${OPTARG};;
esac
done

unset IS_PREVIEW_PIPELINE
unset IS_RELEASE_PIPELINE
# Try and establish what phase of what type of build pipeline we are in
[[ "${PIPELINE_KIND}" == "pullrequest" ]] && [ ! -f .pre-commit-config.yaml ] && IS_PREVIEW_PIPELINE="true" || IS_PREVIEW_PIPELINE="false"
if [[ ${IS_PREVIEW_PIPELINE} == "true" ]] ; then
    echo "Detected Preview pipeline";
fi
[[ "$PIPELINE_KIND" == "release" ]] && [ ! -f .pre-commit-config.yaml ] && IS_RELEASE_PIPELINE="true" || IS_RELEASE_PIPELINE="false"
if [[ ${IS_RELEASE_PIPELINE} == "true" ]] ; then
    echo "Detected Release pipeline";
fi
# Only activate in preview builds or the first stage of a release
if [[ ${IS_PREVIEW_PIPELINE} == "true" ]] || [[ ${IS_RELEASE_PIPELINE} == "true" ]] ; then
    if [[ ${IS_PREVIEW_PIPELINE} == "true" ]] ; then
        if [[ ${SCAN_ON_PREVIEW} == "true" ]] ; then
            echo "Sonarqube is scanning files..."
            if [[ ${BUILDPACK_NAME} == "maven" ]] ; then
                /opt/sonar/bin/sonar-scanner "-Dsonar.host.url=${SONARQUBE_SERVER}" "-Dsonar.projectKey=${JOB_NAME}" "-Dsonar.login=${SONAR_TOKEN}" "-Dsonar.language=java" "-Dsonar.sources=src/main/java" "-Dsonar.java.binaries=target/classes"
            else
                /opt/sonar/bin/sonar-scanner "-Dsonar.host.url=${SONARQUBE_SERVER}" "-Dsonar.projectKey=${JOB_NAME}" "-Dsonar.login=${SONAR_TOKEN}"
            fi
        else
            echo "Sonarqube scanning disabled in preview builds."
        fi
    fi
    if [[ ${IS_RELEASE_PIPELINE} == "true" ]] ; then
        if [[ ${SCAN_ON_RELEASE} == "true" ]] ; then
            echo "Sonarqube is scanning files..."
            if [[ ${BUILDPACK_NAME} == "maven" ]] ; then
                /opt/sonar/bin/sonar-scanner "-Dsonar.host.url=${SONARQUBE_SERVER}" "-Dsonar.projectKey=${JOB_NAME}" "-Dsonar.login=${SONAR_TOKEN}" "-Dsonar.language=java" "-Dsonar.sources=src/main/java" "-Dsonar.java.binaries=target/classes"
            else
                /opt/sonar/bin/sonar-scanner "-Dsonar.host.url=${SONARQUBE_SERVER}" "-Dsonar.projectKey=${JOB_NAME}" "-Dsonar.login=${SONAR_TOKEN}"
            fi
        else
            echo "Sonarqube scanning disabled in release builds."
        fi
    fi
else
    echo "Skipping sonarqube scan"
fi
