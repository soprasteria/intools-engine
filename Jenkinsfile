node{

  cleanWs()

  String gitCredentialsId='Github_DeployKey_IntoolsEngine'

  stage 'Setup'
    properties(
      [[$class: 'jenkins.model.BuildDiscarderProperty',strategy:[$class: 'LogRotator', numToKeepStr: '5', artifactNumToKeepStr: '5']],
      [$class: 'ParametersDefinitionProperty', parameterDefinitions: [[$class: 'StringParameterDefinition', name: 'CUSTOM_VERSION', defaultValue: '',
      description: '[Only used on master branch] Set a custom version if you want to manually set the version tag to a specific value (e.g. bump to a major version). Leave empty if you want to automatically bump the patch version or if you are building develop.']]]]
      )
   // Initializing workspace
   env.CI = true // used to run commands without asking questions to users
    env.GOROOT = tool '1.7.1'
    env.GOPATH = pwd()
    env.PATH = "${GOROOT}/bin:${GOBIN}:${PATH}"
    env.WORKSPACE="${GOPATH}/src/github.com/soprasteria/intools-engine"

  stage 'Checkout'  
      sh 'git config --global credential.helper cache'
      // Checkout the given branch in a sub directory
      checkout([$class: 'GitSCM',
                branches: [[name: '${BRANCH_NAME}']],
                extensions: [[$class: 'RelativeTargetDirectory', relativeTargetDir: env.WORKSPACE], [$class: 'LocalBranch', localBranch: '${BRANCH_NAME}']],
                userRemoteConfigs: [[
                  credentialsId: "${gitCredentialsId}",
                  url: 'git@github.com:soprasteria/intools-engine.git'
              ]]])

      stage 'Compile'
        sh '''
          cd ${WORKSPACE}
          govendor sync -v
          CGO_ENABLED=0 go build -a -installsuffix cgo
        '''
      stage 'Test'
        sh '''
          cd ${WORKSPACE}
          govendor test +local
        '''

    dir(env.WORKSPACE) {
      if (env.BRANCH_NAME == "master") {
        withCredentials([[$class: 'StringBinding', credentialsId: '382b84d3-2bb3-4fca-8d13-7e874c6339a2', variable: 'ARTIFACTORY_URL'], [$class: 'UsernamePasswordBinding', credentialsId: 'cc2089e7-c24c-4048-8311-7376c1bab694', variable: 'ARTIFACTORY_CREDENTIALS']]) {
          stage 'Publish'
            sh '''
              if [ -z "''' + CUSTOM_VERSION + '''"]; then
                version=$(cat version)
              else
                version=''' + CUSTOM_VERSION + '''
              fi

              tarname=intools-engine-$version.tgz
              tar -cvzf $tarname intools-engine

              git tag -af $version -m "version $version"
              git push -f origin $version

              current_dir=`pwd`
              export FILE="$current_dir/$tarname"
              curl -v -u$ARTIFACTORY_CREDENTIALS --data-binary @"${FILE}" -X PUT $ARTIFACTORY_URL/prj-cdk-releases/com/soprasteria/cdk/intools2/intools-engine/$tarname

              majorversion=`echo $version | cut -d '.' -f 1`
              minorversion=`echo $version | cut -d '.' -f 2`
              patchversion=$((`echo $version | cut -d '.' -f 3` + 1))
              newversion="$majorversion.$minorversion.$patchversion"

              echo "Released version" $version
              echo "New version" $newversion

              git fetch
              git checkout develop
              echo $newversion > version
              git status
              git commit -a -m "chore: Bump to version $newversion"
              git status
              git push origin develop
              git checkout master
              git reset --hard origin/master
            '''
        }
      }
  }


  cleanWs()
}
