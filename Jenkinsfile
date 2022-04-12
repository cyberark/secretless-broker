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

    stage('Validate') {
      parallel {
        stage('Changelog') {
          steps { sh './bin/parse-changelog' }
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

    stage('Build and Unit tests') {
      parallel {
        stage('Build Images') {
          steps {
            sh './bin/build'
          }
        }

        stage('Unit tests') {
          steps {
            sh './bin/test_unit'
            sh 'cp ./test/unit-test-output/c.out ./c.out'

            junit 'test/unit-test-output/junit.xml'
            ccCoverage("gocov", "--prefix github.com/cyberark/secretless-broker")
          }
        }
      }
    }

/*
    stage('Scan Secretless') {
      parallel {
        stage('Scan Secretless Image for fixable issues') {
          steps {
            scanAndReport("secretless-broker:latest", "HIGH", false)
          }
        }

        stage('Scan Secretless Image for all issues') {
          steps {
            scanAndReport("secretless-broker:latest", "NONE", true)
          }
        }

        stage('Scan Secretless Quickstart for fixable issues') {
          steps {
            scanAndReport("secretless-broker-quickstart:latest", "HIGH", false)
          }
        }

        stage('Scan Secretless Quickstart for all issues') {
          steps {
            scanAndReport("secretless-broker-quickstart:latest", "NONE", true)
          }
        }

        stage('Scan Secretless RedHat for fixable issues') {
          steps {
            script {
              TAG = sh(returnStdout: true, script: '. bin/build_utils && full_version_tag')
            }
            scanAndReport("secretless-broker-redhat:${TAG}", "HIGH", false)
          }
        }

        stage('Scan Secretless RedHat for all issues') {
          steps {
            script {
              TAG = sh(returnStdout: true, script: '. bin/build_utils && full_version_tag')
            }
            scanAndReport("secretless-broker-redhat:${TAG}", "NONE", true)
          }
        }

        stage('Scan For Security with Gosec') {
          // Gosec only works on branch builds
          when {
            not { tag "v*" }
          }

          steps {
            sh "./bin/check_golang_security -s High -c Medium -b ${env.BRANCH_NAME}"
            junit(allowEmptyResults: true, testResults: 'gosec.output')
          }
        }
      }
    }
    */

    //stage('Integration Tests') {
    //  steps {
    //    script {
    //      def directories = sh (
    //        returnStdout: true,
    //        // We run the 'find' directive first on all directories with test files, then run a 'find' directive
    //        // to make sure they also contain start files. We then take the dirname, and basename respectively.
    //        script:
    //        '''
    //        find $(find ./test -name test) -name 'start' -exec dirname {} \\; | xargs -n1 basename
    //        '''
    //      ).trim().split()
//
//          def integrationStages = [:]
//
//          // Create an integration test stage for each directory we collected previously.
//          // We want to be sure to skip any tests, such as keychain tests, that can only be run manually.
//          directories.each { name ->
//            if (name == "keychain") return
//
//            integrationStages["Integration: ${name}"] = {
//              sh "./bin/run_integration ${name}"
//            }
//          }
//
//          parallel integrationStages
//        }
//        junit "**/test/**/junit.xml"
//      }
//    }

/*
    stage('Combine Integration and Unit Test Coverage') {
      steps {
        sh "./bin/merge_integration_coverage"
        archiveArtifacts 'test/test-coverage/integ-and-ut-cover.html'
      }
    }

    stage('Cobertura') {
      steps {
        cobertura autoUpdateHealth: false, autoUpdateStability: false, coberturaReportFile: 'test/test-coverage/coverage.xml', conditionalCoverageTargets: '50, 0, 0', failUnhealthy: false, failUnstable: false, lineCoverageTargets: '50, 0, 0', maxNumberOfBuilds: 0, methodCoverageTargets: '50, 0, 0', onlyStable: false, sourceEncoding: 'ASCII', zoomCoverageChart: false
      }
    }

    stage('Functional Tests') {
      parallel {
        stage('Quick start') {
          steps {
            sh './bin/test_demo'
          }
        }

        stage('K8s Demo') {
          steps {
            sh 'summon -f ./k8s-ci/secrets.yml ./k8s-ci/test demos/k8s-demo'
          }
        }

        stage('CRD test') {
          steps {
            sh 'summon -f ./k8s-ci/secrets.yml ./k8s-ci/test k8s-ci/k8s_crds'
          }
        }
      }
    }
    */

    stage('Push Images Internally') {
      steps {
        sh './bin/publish --internal'
      }
    }

    stage('Build Release Artifacts') {
      //when {
      //  branch 'main'
      //}

      steps {
        sh './bin/build_release --snapshot'
        archiveArtifacts 'dist/goreleaser/'
      }
    }

    stage('Release') {
      when {
        expression {
          MODE == "RELEASE"
        }
      }
      steps {
        release { billOfMaterialsDirectory, assetDirectory, toolsDirectory ->
          // Publish release artifacts to all the appropriate locations
          // Copy any artifacts to assetDirectory to attach them to the Github release

          //    // Create Go application SBOM using the go.mod version for the golang container image
          sh """go-bom --tools "${toolsDirectory}" --go-mod ./go.mod --image "golang" --main "cmd/secretless-broker/" --output "${billOfMaterialsDirectory}/go-app-bom.json" """
          //    // Create Go module SBOM
          sh """go-bom --tools "${toolsDirectory}" --go-mod ./go.mod --image "golang" --output "${billOfMaterialsDirectory}/go-mod-bom.json" """
          sh 'summon -e production ./bin/publish --edge'
        }
      }
    }

/*
    // Must com after release block as it relies on the pre-release version
    stage('Create draft release') {
      steps {
        dir('./pristine-checkout') {
          // Go releaser requires a pristine checkout
          checkout scm
          sh 'git submodule update --init --recursive'
          // Create draft release
          sh "summon --yaml 'GITHUB_TOKEN: !var github/users/conjur-jenkins/api-token' ./bin/build_release"
          archiveArtifacts 'dist/goreleaser/'
        }
      }
    }

    stage('Fix Website Flags (staging)') {
      when {
        branch 'staging'
      }
      steps {
        sh 'sed -i "s#^is_maintenance_mode:.*#is_maintenance_mode: false#" docs/_config.yml'
      }
    }

    stage('Build Website') {
      steps {
        sh './bin/build_website'
      }
    }

    stage('Check Links') {
      steps {
        sh './bin/check_website_links'
      }
    }

    stage('Publish') {
      parallel {
        stage('Publish Website (staging)') {
          when {
            branch 'staging'
          }
          steps {
            sh 'summon -e staging bin/publish_website'
            archiveArtifacts 'docs/_site/'
          }
        }

        stage('Publish Website (production)') {
          when {
            branch 'main'
          }
          steps {
            sh 'summon -e production bin/publish_website'
            archiveArtifacts 'docs/_site/'
          }
        }
      }
    }
  */
  }

  post {
    always {
      cleanupAndNotify(currentBuild.currentResult)
    }
  }
}
