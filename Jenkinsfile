#!/usr/bin/env groovy
@Library("product-pipelines-shared-library") _

// Automated release, promotion and dependencies
properties([
  // Include the automated release parameters for the build
  release.addParams(),
  // Dependencies of the project that should trigger builds
  dependencies([
    'conjur-enterprise/conjur-opentelemetry-tracer',
    'conjur-enterprise/conjur-api-go',
    'conjur-enterprise/conjur-authn-k8s-client',
    'conjur-enterprise/summon'
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
    infrapool.agentSh """
      export PATH="release-tools/bin:${PATH}"
      docker pull registry.tld/secretless-broker:${sourceVersion}
      docker pull registry.tld/secretless-broker-quickstart:${sourceVersion}
      docker pull registry.tld/secretless-broker-redhat:${sourceVersion}
      #Promote source version to target version.
      summon -e common ./bin/publish --promote --source ${sourceVersion} --target ${targetVersion}
    """
  }

  // Copy Github Enterprise release to Github
  release.copyEnterpriseRelease(params.VERSION_TO_PROMOTE)
  return
}

pipeline {
  agent { label 'conjur-enterprise-common-agent' }

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

    stage('Get InfraPool ExecutorV2 Agent') {
      steps {
        script {
          // Request ExecutorV2 agents for 1 hour(s)
          INFRAPOOL_EXECUTORV2_AGENTS = getInfraPoolAgent(type: "ExecutorV2", quantity: 1, duration: 5)
          INFRAPOOL_EXECUTORV2_AGENT_0 = INFRAPOOL_EXECUTORV2_AGENTS[0]
        }
      }
    }

    stage('Validate Changelog') {
      steps {
        script {
          infraPoolConnect(INFRAPOOL_EXECUTORV2_AGENT_0) { infrapool ->
            infrapool.agentSh './bin/parse-changelog'
          }
        }
      }
    }

    // Generates a VERSION file based on the current build number and latest version in CHANGELOG.md
    stage('Validate Changelog and set version') {
      steps {
        script {
          infraPoolConnect(INFRAPOOL_EXECUTORV2_AGENT_0) { infrapool ->
            updateVersion(infrapool, "CHANGELOG.md", "${BUILD_NUMBER}")
          }
        }
      }
    }

    stage('Update Submodules') {
      steps {
        script {
          infraPoolConnect(INFRAPOOL_EXECUTORV2_AGENT_0) { infrapool ->
            infrapool.agentSh 'git submodule update --init --recursive'
          }
        }
      }
    }

    stage('Get latest upstream dependencies') {
      steps {
        script {
          infraPoolConnect(INFRAPOOL_EXECUTORV2_AGENT_0) { infrapool ->
            updateGoDependencies(infrapool, "${WORKSPACE}/go.mod")
          }
        }
      }
    }

    stage('Build and Unit tests') {
      parallel {
        stage('Build Images') {
          steps {
            script {
              infraPoolConnect(INFRAPOOL_EXECUTORV2_AGENT_0) { infrapool ->
                infrapool.agentSh './bin/build'
              }
            }
          }
        }

        stage('Unit tests') {
          steps {
            script {
              infraPoolConnect(INFRAPOOL_EXECUTORV2_AGENT_0) { infrapool ->
                infrapool.agentSh './bin/test_unit'
              }
            }
          }
          post {
            always {
              script {
                infraPoolConnect(INFRAPOOL_EXECUTORV2_AGENT_0) { infrapool ->
                  infrapool.agentSh './bin/coverage'
                  infrapool.agentSh 'cp ./test/unit-test-output/c.out ./c.out'
                  infrapool.agentStash name: 'junit-report', includes: 'test/unit-test-output/junit.xml'
                }
              }

              unstash 'junit-report'
              junit 'test/unit-test-output/junit.xml'
            }
          }
        }
      }
    }

    stage('Scan Secretless') {
      parallel {
        stage('Scan Secretless Image for fixable issues') {
          steps {
            script {
              infraPoolConnect(INFRAPOOL_EXECUTORV2_AGENT_0) { infrapool ->
                scanAndReport(infrapool, "secretless-broker:latest", "HIGH", false)
              }
            }
          }
        }

        stage('Scan Secretless Image for all issues') {
          steps {
           script {
              infraPoolConnect(INFRAPOOL_EXECUTORV2_AGENT_0) { infrapool ->
                scanAndReport(infrapool, "secretless-broker:latest", "NONE", true)
              }
            }
          }
        }

        stage('Scan Secretless Quickstart for fixable issues') {
          steps {
           script {
              infraPoolConnect(INFRAPOOL_EXECUTORV2_AGENT_0) { infrapool ->
                scanAndReport(infrapool, "secretless-broker-quickstart:latest", "HIGH", false)
              }
            }
          }
        }

        stage('Scan Secretless Quickstart for all issues') {
          steps {
           script {
              infraPoolConnect(INFRAPOOL_EXECUTORV2_AGENT_0) { infrapool ->
                scanAndReport(infrapool, "secretless-broker-quickstart:latest", "NONE", true)
              }
            }
          }
        }

        stage('Scan Secretless RedHat for fixable issues') {
          steps {
            script {
              infraPoolConnect(INFRAPOOL_EXECUTORV2_AGENT_0) { infrapool ->
                TAG = infrapool.agentSh(returnStdout: true, script: '. bin/build_utils && full_version_tag')
                scanAndReport(infrapool, "secretless-broker-redhat:${TAG}", "HIGH", false)
              }
            }
          }
        }

        stage('Scan Secretless RedHat for all issues') {
          steps {
            script {
              infraPoolConnect(INFRAPOOL_EXECUTORV2_AGENT_0) { infrapool ->
                TAG = infrapool.agentSh(returnStdout: true, script: '. bin/build_utils && full_version_tag')
                scanAndReport(infrapool, "secretless-broker-redhat:${TAG}", "NONE", true)
              }
            }
          }
        }

        stage('Scan For Security with Gosec') {
          //Gosec only works on branch builds
          when {
            not { buildingTag() }
          }
          steps {
            script {
              infraPoolConnect(INFRAPOOL_EXECUTORV2_AGENT_0) { infrapool ->
                infrapool.agentGet from: "${WORKSPACE}", to: "${WORKSPACE}"
              }
              sh "./bin/check_golang_security -s High -c Medium -b ${env.BRANCH_NAME}"
            }
            junit(allowEmptyResults: true, testResults: 'gosec.output')
          }
        }
      }
    }

    stage('Integration Tests') {
      steps {
        script {
          infraPoolConnect(INFRAPOOL_EXECUTORV2_AGENT_0) { infrapool ->
            def directories = infrapool.agentSh (
              returnStdout: true,
              // We run the 'find' directive first on all directories with test files, then run a 'find' directive
              // to make sure they also contain start files. We then take the dirname, and basename respectively.
              script:
              '''
              find $(find ./test -name test) -name 'start' -exec dirname {} \\; | xargs -n1 basename
              '''
            ).trim().split()
            def integrationStages = [:]


            // Create an integration test stage for each directory we collected previously.
            // We want to be sure to skip any tests, such as keychain tests, that can only be run manually.
            directories.each { name ->
              if (name == "keychain") return

              integrationStages["Integration: ${name}"] = {
                infrapool.agentSh "./bin/run_integration ${name}"
                infrapool.agentStash name: 'integration-junit-report', includes: '**/test/**/junit.xml'
              }
            }

            parallel integrationStages
          }
        }
        unstash 'integration-junit-report'
        junit "**/test/**/junit.xml"
      }
    }

    stage('Combine Integration and Unit Test Coverage') {
      steps {
        script {
          infraPoolConnect(INFRAPOOL_EXECUTORV2_AGENT_0) { infrapool ->
            infrapool.agentSh "./bin/merge_integration_coverage"
            infrapool.agentArchiveArtifacts artifacts: 'test/test-coverage/integ-and-ut-cover.html'
          }
        }
      }
    }

    stage('Cobertura') {
      steps {
        script {
          infraPoolConnect(INFRAPOOL_EXECUTORV2_AGENT_0) { infrapool ->
            cobertura(
              autoUpdateHealth: false,
              autoUpdateStability: false,
              coberturaReportFile: 'test/test-coverage/coverage.xml',
              conditionalCoverageTargets: '70, 0, 70',
              failUnhealthy: true,
              failUnstable: true,
              lineCoverageTargets: '70, 70, 70',
              maxNumberOfBuilds: 0,
              methodCoverageTargets: '70, 0, 70',
              onlyStable: false,
              sourceEncoding: 'ASCII',
              zoomCoverageChart: false
            )
            infrapool.agentSh 'cp test/test-coverage/integ-and-ut-cover.out ./c.out'
            infrapool.agentStash name: 'coverage-report', includes: 'test/test-coverage/coverage.xml'
          }
          codacy action: 'reportCoverage', filePath: "test/test-coverage/coverage.xml"
        }
      }
    }

    stage('Functional Tests') {
      parallel {
        stage('Quick start') {
          steps {
            script {
              infraPoolConnect(INFRAPOOL_EXECUTORV2_AGENT_0) { infrapool ->
                infrapool.agentSh './bin/test_demo'
              }
            }
          }
        }

        stage('K8s Demo') {
          steps {
            script {
              infraPoolConnect(INFRAPOOL_EXECUTORV2_AGENT_0) { infrapool ->
                infrapool.agentSh 'summon -f ./k8s-ci/secrets.yml ./k8s-ci/test demos/k8s-demo'
              }
            }
          }
        }

        stage('CRD test') {
          steps {
            script {
              infraPoolConnect(INFRAPOOL_EXECUTORV2_AGENT_0) { infrapool ->
                infrapool.agentSh 'summon -f ./k8s-ci/secrets.yml ./k8s-ci/test k8s-ci/k8s_crds'
              }
            }
          }
        }
      }
    }

    stage('Push Images Internally') {
      steps {
        script {
          infraPoolConnect(INFRAPOOL_EXECUTORV2_AGENT_0) { infrapool ->
            infrapool.agentSh'./bin/publish --internal'
          }
        }
      }
    }

    stage('Build Release Artifacts') {
      when {
        expression {
          MODE == "RELEASE"
        }
      }

      steps {
        script {
          infraPoolConnect(INFRAPOOL_EXECUTORV2_AGENT_0) { infrapool ->
            infrapool.agentSh './bin/build_release --snapshot'
            infrapool.agentArchiveArtifacts artifacts: 'dist/goreleaser/'
          }
        }
      }
    }

    stage('Create Release Assets') {
      when {
        expression {
          MODE == "RELEASE"
        }
      }
      steps {
        script {
          infraPoolConnect(INFRAPOOL_EXECUTORV2_AGENT_0) { infrapool ->
            dir('./pristine-checkout') {
              // Go releaser requires a pristine checkout
              checkout scm
              infrapool.agentPut from: """${WORKSPACE}/pristine-checkout""", to: """$WORKSPACE/"""
              infrapool.agentSh 'git submodule update --init --recursive'
              // Create release packages without releasing to Github
              infrapool.agentSh "./bin/build_release --skip-validate"
              infrapool.agentArchiveArtifacts artifacts: 'dist/goreleaser/'
            }
          }
        }
      }
    }

    stage('Release') {
      when {
        expression {
          MODE == "RELEASE"
        }
      }
      steps {
        script {
          infraPoolConnect(INFRAPOOL_EXECUTORV2_AGENT_0) { infrapool ->
            release(infrapool) { billOfMaterialsDirectory, assetDirectory, toolsDirectory ->
              // Publish release artifacts to all the appropriate locations
              // Copy any artifacts to assetDirectory to attach them to the Github release

              // Copy assets to be published in Github release.
              infrapool.agentSh "./bin/copy_release_assets ${assetDirectory}"

              // Create Go application SBOM using the go.mod version for the golang container image
              infrapool.agentSh """export PATH="${toolsDirectory}/bin:${PATH}" && go-bom --tools "${toolsDirectory}" --go-mod ./go.mod --image "golang" --main "cmd/secretless-broker/" --output "${billOfMaterialsDirectory}/go-app-bom.json" """
              // Create Go module SBOM
              infrapool.agentSh """export PATH="${toolsDirectory}/bin:${PATH}" && go-bom --tools "${toolsDirectory}" --go-mod ./go.mod --image "golang" --output "${billOfMaterialsDirectory}/go-mod-bom.json" """
              infrapool.agentSh """export PATH="${toolsDirectory}/bin:${PATH}" && summon -e production ./bin/publish --edge"""
            }
          }
        }
      }
    }

    stage('Fix Website Flags (staging)') {
      when {
        branch 'staging'
      }
      steps {
        script {
          infraPoolConnect(INFRAPOOL_EXECUTORV2_AGENT_0) { infrapool ->
            infrapool.agentSh 'sed -i "s#^is_maintenance_mode:.*#is_maintenance_mode: false#" docs/_config.yml'
          }
        }
      }
    }

    stage('Build Website') {
      steps {
        script {
          infraPoolConnect(INFRAPOOL_EXECUTORV2_AGENT_0) { infrapool ->
            infrapool.agentSh './bin/build_website'
          }
        }
      }
    }

    stage('Check Links') {
      steps {
        script {
          infraPoolConnect(INFRAPOOL_EXECUTORV2_AGENT_0) { infrapool ->
            infrapool.agentSh './bin/check_website_links'
          }
        }
      }
    }

    stage('Publish') {
      parallel {
        stage('Publish Website (staging)') {
          when {
            branch 'staging'
          }
          steps {
            script {
              infraPoolConnect(INFRAPOOL_EXECUTORV2_AGENT_0) { infrapool ->
                infrapool.agentSh 'summon -e staging bin/publish_website'
                infrapool.agentArchiveArtifacts artifacts: 'docs/_site/'
              }
            }
          }
        }

        stage('Publish Website (production)') {
          when {
            branch 'main'
          }
          steps {
            script {
              infraPoolConnect(INFRAPOOL_EXECUTORV2_AGENT_0) { infrapool ->
                infrapool.agentSh 'summon -e production bin/publish_website'
                infrapool.agentArchiveArtifacts artifacts: 'docs/_site/'
              }
            }
          }
        }
      }
    }
  }

  post {
    always {
      releaseInfraPoolAgent(".infrapool/release_agents")
    }
  }
}
