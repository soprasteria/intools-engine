node{

  cleanWs()

  String gitCredentialsId='Github_DeployKey_IntoolsEngine'

  stage('Setup'){
    properties(
      [[$class: 'jenkins.model.BuildDiscarderProperty',strategy:[$class: 'LogRotator', numToKeepStr: '5', artifactNumToKeepStr: '5']],
      [$class: 'ParametersDefinitionProperty', parameterDefinitions: [[$class: 'StringParameterDefinition', name: 'CUSTOM_VERSION', defaultValue: '',
      description: '[Only used on master branch] Set a custom version if you want to manually set the version tag to a specific value (e.g. bump to a major version). Leave empty if you want to automatically bump the patch version or if you are building develop.']]]]
      )
   // Initializing workspace
   env.CI = true // used to run commands without asking questions to users
    env.GOROOT = tool '1.7.1'
    env.GOPATH = pwd()
    env.GOBIN = "${GOPATH}/bin"
    env.PATH = "${GOROOT}/bin:${GOBIN}:${PATH}"
    env.WORKSPACE="${GOPATH}/src/github.com/soprasteria/intools-engine"
    
    // get Govendor
    sh '''
       go get -u github.com/kardianos/govendor
    '''
    
  }
  

  stage ('Checkout'){
      sh 'git config --global credential.helper cache'
      // Checkout the given branch in a sub directory
      checkout([$class: 'GitSCM',
                branches: [[name: '${BRANCH_NAME}']],
                extensions: [[$class: 'RelativeTargetDirectory', relativeTargetDir: env.WORKSPACE], [$class: 'LocalBranch', localBranch: '${BRANCH_NAME}']],
                userRemoteConfigs: [[
                  credentialsId: "${gitCredentialsId}",
                  url: 'git@github.com:soprasteria/intools-engine.git'
              ]]])
  }
  
  stage('Compile'){
        sh '''
          cd "${WORKSPACE}"
          govendor sync -v
          CGO_ENABLED=0 go build -a -installsuffix cgo
        '''
   }
  stage ('Test') {
        sh '''
          cd "${WORKSPACE}"
          govendor test +local
        '''
  }
  

  if (env.BRANCH_NAME == "master") {
    dir(env.WORKSPACE) {
        def version = (CUSTOM_VERSION?.trim())?CUSTOM_VERSION:sh(returnStdout: true, script:'cat version')
        version = version.trim()
        withEnv(["VERSION=${version}"]){
          stage ('Publish'){
            sh '''
              tarname=intools-engine-${VERSION}.tgz
              tar -cvzf $tarname intools-engine
            '''

            sshagent(credentials: ["${gitCredentialsId}"]){
              sh '''
                git tag -af ${VERSION} -m "version ${VERSION}"
                git push -f origin ${VERSION}
              '''
            }
            withCredentials([[$class: 'StringBinding', credentialsId: '382b84d3-2bb3-4fca-8d13-7e874c6339a2', variable: 'ARTIFACTORY_URL'], [$class: 'UsernamePasswordBinding', credentialsId: 'cc2089e7-c24c-4048-8311-7376c1bab694', variable: 'ARTIFACTORY_CREDENTIALS']]) {
              sh '''
                current_dir=`pwd`
                export FILE="$current_dir/$tarname"
                curl -v -u$ARTIFACTORY_CREDENTIALS --data-binary @"${FILE}" -X PUT $ARTIFACTORY_URL/prj-cdk-releases/com/soprasteria/cdk/intools2/intools-engine/$tarname
              '''
            }
          }
          stage ('Update Version on SCM'){
            sshagent(credentials: ["${gitCredentialsId}"]){
              sh '''
                majorversion=`echo ${VERSION} | cut -d '.' -f 1`
                minorversion=`echo ${VERSION} | cut -d '.' -f 2`
                patchversion=$((`echo ${VERSION} | cut -d '.' -f 3` + 1))
                newversion="$majorversion.$minorversion.$patchversion"

                echo "Released version" ${VERSION}
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
  }


  cleanWs()
}
