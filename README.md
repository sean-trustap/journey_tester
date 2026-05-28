# Journey Tester

## About

The journey tester is a tool that runs collections of journeys. A journey is an
end-to-end test made of HTTP requests and expected responses that simulates a
user's journey through an API.

## Usage

### Installation

Install the CLI tool:

    go install github.com/trustap/journey_tester/cmd/journey_tester@latest

### Getting started

The tests can have test data, such as `Users` and `FeeConfig`. Users take time
to create and do not need to be created every time if you are running the tests
locally for development.

To just create test data:

    journey_tester -create -skip-testing

To run the tests:

    journey_tester

To just delete test data:

    journey_tester -delete -skip-testing

## Discovery

The tool discovers journey functions, data journeys, and setup steps from
your `test/` directory:

- Journey functions: `func MyTest() *journey_tester.Journey`
- Data journeys: `var DataJourneys = ...`
- Setup steps: `var BeforeEachTest = []*journey_tester.Step{...}`

## Options

Journeys can be associated with groups. You can control which groups are run
using the `groups` parameter:

    journey_tester -groups user

You can specify multiple groups:

    journey_tester -groups user,transaction

You can also specify exactly which tests should be run using the `run`
parameter, with a folder path:

    journey_tester -run tests/user

Or a file name:

    journey_tester -run transaction_happy_path.go

## Special Groups

* `manual`: If a journey is tagged with this group, it will not be run
  automatically. You must use the `run` parameter and specify the journey's
  filename to run the test.
* `non_parallel`: If a journey is tagged with this group, it means it must be
  run sequentially and will have undefined behaviour if run in parallel with
  other journeys.

## Development

### Build environment

The build environment for this project is defined in `build.Dockerfile`. These
steps can be used to replicate the build environment locally. Alternatively, to
make use of the Dockerised build environment, the following can be used:

    bash scripts/with_build_env.sh --dev bash

This command will start a new Bash session within the build environment, with
the local project directory and local user mounted inside the build environment,
which can be used for local development.

### Building

#### With Docker

If Docker is installed then the tests can be run as follows:

    bash scripts/with_build_env.sh just check

Because they're heavier, journey tests are run separately to regular checks. To
run them, populate `test/envs/vars.env` from `test/envs/vars.sample.env`, and
run the following:

    bash scripts/with_build_env.sh just check_jrns

#### Without Docker

The instructions in `build.Dockerfile` can be followed to prepare your local
environment for building the project. With the local environment set up, the
project can be tested locally by running `just check`.

#### Outputs

Multiple directories are generated to the `target` directory.

* `artfs` contains build artefacts that can be reused by other projects. For
  example, a library binary or command executable, which can be added to a
  Docker image. These are generally copied to an artefact server to represent a
  given version of the codebase.
* `gen` contains artefacts that are used to build or test other artefacts, but
  which themselves aren't reusable. For example, this directory may contain a
  mock server for integration testing, or a code generator needed to create
  source files for the final build. These shouldn't be copied to an artefact
  server. It's safe to delete this directory, though generally the artefacts in
  this directory are useful to keep in order to speed up future builds.
* `tmp` contains temporary data that may be used by some build processes. For
  example, tests may require a directory where test data or reports can be
  written. These will generally be deleted and rebuilt each time the associated
  build process runs, so they can be deleted without problems.

### Project layout

This project is mainly laid out according to
<https://github.com/golang-standards/project-layout>, and conforms to the
standards laid out in
<https://trustap.atlassian.net/wiki/spaces/ENGINEERING/pages/427950081/Trustap+Code+Style>.
