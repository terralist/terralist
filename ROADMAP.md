# Terralist Road Map

### What can I find in this document?

This document providers a list of Terralist's objectives and serves as a guideline for the project contributors, to better understand what are the future plans of this project.

If you are not a contributor, you can still use this document to follow the Terralist progress.

## v0.5.0

+ Rework of the API Keys & Tokens
  + Decouple API Keys from Authorities;
  + Add ACL for API Keys:
    + Read/Write access per each Authority;
    + Ability to create new Authorities;
    + Ability to create new API Keys;
  + JWT Tokens issued for Terraform-CLI should only be granted with read access to all modules/providers;
  + JWT Tokens issued for Terralist Web UI should also be restricted (depending on a given configuration);

## v0.6.0

+ Integration testing for all public APIs & for each supported database;
+ End-to-end testing;
+ Unit testing for repositories; 

## v0.7.0

+ Modules & providers documentation from the web interface;

## v1.0.0

_Note_: The `v1.0.0` tag will be released if no bug is reported (on a given period of time) and all features before this version have been implemented successfully.

## Other features

The following list contains other features that will be implemented in a version prior `v1.0.0` release, once all requirements for them have been established.

+ Support for modules upload via files;
+ Support for providers upload via files;
+ Integration with GoReleaser for providers upload;
+ Official Helm Chart;
+ GitHub Action for publishing a module;
+ GitHub Action for publishing a provider;
+ Terraform Provider (for creating API Keys & Authorities);

_Note_ This is not an exclusive list, other features can still be requested by [opening an issue](https://github.com/valentindeaconu/terralist/issues).

## Other organizational changes

+ Add a development environment (dev-container);
+ Create a community Slack/Discord channel;
+ Migrate Terralist to an organization;
+ Release Terralist documentation to terralist.io;