# Terralist Road Map

### What can I find in this document?

This document providers a list of Terralist's objectives and serves as a guideline for the project contributors, to better understand what are the future plans of this project.

If you are not a contributor, you can still use this document to follow the Terralist progress and more especially, in this period, how long it will take for the project first release.

## Required features for the release

1. Web interface protected by OAUTH.
2. Ability to manage (create/update/delete) a single authority per account from the web interface.
3. Ability to create multiple API keys from the web interface.
4. Ability to delete an API key from the web interface.
5. S3 backend support for providers.
6. Getting started documentation.
7. Configuration documentation.
8. Unit testing for core functionalities.

## Release

1. GitHub Action workflow to handle the release.
2. Generate and upload binary files to a release.
3. Upload the containerized version to GitHub Container Registry.
4. Create and maintain a changelog.

## Immediately post-release goals

1. Default AWS credential chain support. Tracking issue: [#28](https://github.com/valentindeaconu/terralist/issues/28).
2. Unit testing for all components.
3. Web interface styling.
4. Live development environment, running latest release version - will also serve as demo.

## Future plans

1. Ability to read modules & providers documentation from the web interface.
2. Replace PostgreSQL with a lighter database.
3. Ability to manage multiple authorities.
4. Google OAUTH 2.0 provider.
