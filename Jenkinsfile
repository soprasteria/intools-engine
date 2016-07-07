node{
  stage 'Setup'
    // Log rotation
    properties([[$class: 'jenkins.model.BuildDiscarderProperty', strategy:[$class: 'LogRotator', numToKeepStr: '5', artifactNumToKeepStr: '5']]])

    def goTool = tool 'GO-1.5.3'
    // clean build
    sh 'rm -rf src'
    env.INTOOLS_BUILD = "src/github.com/soprasteria/intools-engine"
    env.PATH = "$goTool/bin:${env.PATH}"
    env.GOROOT = "$goTool"
    env.GOPATH = pwd()
  stage 'Checkout'
    // Checkout the given branch in a sub directory
    checkout([$class: 'GitSCM',
              branches: [[name: '${BRANCH_NAME}']],
              extensions: [[$class: 'RelativeTargetDirectory', relativeTargetDir: env.INTOOLS_BUILD], [$class: 'LocalBranch', localBranch: '${BRANCH_NAME}']],
              userRemoteConfigs: [[url: 'https://github.com/soprasteria/intools-engine.git']]])
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
    stage 'Package'
      // TODO package application
      if (env.BRANCH_NAME == "master") {
        stage 'Publish'
        // TODO upload to artifactory
        // TODO bump version in develop branch
      }
  }
}
