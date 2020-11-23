---
name: Release
about: Create a new release of the project
title: 'Create a new v{insert version} release of Secretless'
labels: component/secretless-broker, kind/release
assignees: ''

---

AC:
- [ ] The CHANGELOG in the version bump PR has been reviewed to ensure it meets
      our [standards](https://github.com/cyberark/community/blob/master/Conjur/CONTRIBUTING.md#changelog-guidelines)
      and captures the changes in the release.
- [ ] The manual tests have been run for the code on master in the version bump
      PR.
- [ ] There is a new Secretless tag for the set of changes on master.
- [ ] There is a github prerelease for the tag that includes a copy/paste of the
      change log and includes attached artifacts produced in the tag pipeline by
      goreleaser.
- [ ] The changeset has been reviewed with the Quality Architect to determine if
      any performance tests (short- or long-term) are required for this set of
      changes.
- [ ] Any open docs issues for the changes included in the release have been
      made and are ready to be published.
- [ ] If the docs are published and the performance tests aren't needed or are
      complete and passing, then the github release is moved from a prerelease
      to a full release, and a new `stable` image is produced as a copy of the
      new tag and pushed to DockerHub.
