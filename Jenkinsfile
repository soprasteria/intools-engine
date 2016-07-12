node{
  stage 'Setup'
    properties(
      [[$class: 'jenkins.model.BuildDiscarderProperty',strategy:[$class: 'LogRotator', numToKeepStr: '5', artifactNumToKeepStr: '5']],
      [$class: 'ParametersDefinitionProperty', parameterDefinitions: [[$class: 'StringParameterDefinition', name: 'CUSTOM_VERSION', defaultValue: '',
      description: 'Set a custom version if you want to manually set the version tag to a specific value (e.g. bump to a major version). Leave empty if you want to automatically bump the patch version.']]]]
      )

    def goTool = tool 'GO-1.5.3'
    // clean build
    sh 'rm -rf src'
    env.INTOOLS_BUILD = "src/github.com/soprasteria/intools-engine"
    env.PATH = "$goTool/bin:${env.PATH}"
    env.GOROOT = "$goTool"
    env.GOPATH = pwd()
  stage 'Checkout'
    withCredentials([[$class: 'UsernamePasswordMultiBinding', credentialsId: '9ec20a0a-6264-4217-8ac0-11df115c70cc', passwordVariable: 'GITHUB_ACCESS_TOKEN', usernameVariable: 'GITHUB_LOGIN']]) {
      sh 'git config --global credential.helper cache'
      // Checkout the given branch in a sub directory
      checkout([$class: 'GitSCM',
                branches: [[name: '${BRANCH_NAME}']],
                extensions: [[$class: 'RelativeTargetDirectory', relativeTargetDir: 'src/github.com/soprasteria/intools-engine'], [$class: 'LocalBranch', localBranch: '${BRANCH_NAME}']],
                userRemoteConfigs: [[url: 'https://${GITHUB_LOGIN}:${GITHUB_ACCESS_TOKEN}@github.com/soprasteria/intools-engine.git']]])
    }
  dir(env.INTOOLS_BUILD){
    stage 'Compile'
      sh '''
        go get -v
        go build
      '''
    stage 'Test'
      sh '''
        go get -t ./...
        go test ./...
      '''

      if (env.BRANCH_NAME == "master") {
        withCredentials([[$class: 'StringBinding', credentialsId: '382b84d3-2bb3-4fca-8d13-7e874c6339a2', variable: 'ARTIFACTORY_URL'], [$class: 'UsernamePasswordBinding', credentialsId: 'cc2089e7-c24c-4048-8311-7376c1bab694', variable: 'ARTIFACTORY_CREDENTIALS']]) {
          stage 'Publish'
            sh '''
              if [ -Z $CUSTOM_VERSION ]; then
                version=$(cat version)
              else
                version=$CUSTOM_VERSION
              fi

              tarname=intools-engine-$version.tgz
              tar -cvzf $tarname intools-engine

              git tag -af $version -m "version $version"
              git push -f origin $version

              current_dir=`pwd`
              export FILE="$current_dir/$tarname"
              curl -v -u$ARTIFACTORY_CREDENTIALS --data-binary @${FILE} -X PUT $ARTIFACTORY_URL/prj-cdk-releases/com/soprasteria/cdk/intools2/intools-engine/$tarname

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
}
