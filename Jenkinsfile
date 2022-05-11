#!/usr/bin/env groovy

// Automated release, promotion and dependencies
properties([
  // Include the automated release parameters for the build
  release.addParams(),
  // Dependencies of the project that should trigger builds
  dependencies([
    'cyberark/conjur-opentelemetry-tracer',
    'cyberark/conjur-api-go',
    'cyberark/conjur-authn-k8s-client',
    'cyberark/summon'
  ])
])

// Performs release promotion.  No other stages will be run
if (params.MODE == "PROMOTE") {
  release.promote(params.VERSION_TO_PROMOTE) { sourceVersion, targetVersion, assetDirectory ->
    // Any assets from sourceVersion Github release are available in assetDirectory
    // Any version number updates from sourceVersion to targetVersion occur here
    // Any publishing of targetVersion artifacts occur here
    // Anything added to assetDirectory will be attached to the Github Release

    // Pull existing images from internal registry in order to promote
    sh "docker pull registry.tld/secretless-broker:${sourceVersion}"
    sh "docker pull registry.tld/secretless-broker-quickstart:${sourceVersion}"
    sh "docker pull registry.tld/secretless-broker-redhat:${sourceVersion}"
    // Promote source version to target version.
    sh "summon ./bin/publish --promote --source ${sourceVersion} --target ${targetVersion}"
  }
  return
}

pipeline {
  agent { label 'executor-v2' }

  options {
    timestamps()
    buildDiscarder(logRotator(numToKeepStr: '30'))
  }

  triggers {
    cron(getDailyCronString())
  }

  environment {
    // Sets the MODE to the specified or autocalculated value as appropriate
    MODE = release.canonicalizeMode()
  }

  stages {
    // Aborts any builds triggered by another project that wouldn't include any changes
    stage ("Skip build if triggering job didn't create a release") {
      when {
        expression {
          MODE == "SKIP"
        }
      }
      steps {
        script {
          currentBuild.result = 'ABORTED'
          error("Aborting build because this build was triggered from upstream, but no release was built")
        }
      }
    }
    
    // Generates a VERSION file based on the current build number and latest version in CHANGELOG.md
    stage('Validate Changelog and set version') {
      steps {
        updateVersion("CHANGELOG.md", "${BUILD_NUMBER}")
      }
    }

    stage('Update Submodules') {
      steps {
        sh 'git submodule update --init --recursive'
      }
    }

    stage('Get latest upstream dependencies') {
      steps {
        updateGoDependencies("${WORKSPACE}/go.mod")
      }
    }

    stage('Create Release Assets') {
      steps {
        dir('./pristine-checkout') {
          // Go releaser requires a pristine checkout
          checkout scm
          sh 'git submodule update --init --recursive'
          // Create release packages without releasing to Github
          sh "./bin/build_release --skip-validate"
          archiveArtifacts 'dist/goreleaser/'
        }
      }
    }

    stage('Release') {
      steps {
        release { billOfMaterialsDirectory, assetDirectory, toolsDirectory ->
          // Publish release artifacts to all the appropriate locations
          // Copy any artifacts to assetDirectory to attach them to the Github release

          // Copy assets to be published in Github release.
          sh "./bin/copy_release_assets ${assetDirectory}"
          sh "exit 1"

          // Create Go application SBOM using the go.mod version for the golang container image
          sh """go-bom --tools "${toolsDirectory}" --go-mod ./go.mod --image "golang" --main "cmd/secretless-broker/" --output "${billOfMaterialsDirectory}/go-app-bom.json" """
          // Create Go module SBOM
          sh """go-bom --tools "${toolsDirectory}" --go-mod ./go.mod --image "golang" --output "${billOfMaterialsDirectory}/go-mod-bom.json" """
          sh 'summon -e production ./bin/publish --edge'
        }
      }
    }
  }

  post {
    always {
      cleanupAndNotify(currentBuild.currentResult)
    }
  }
}
