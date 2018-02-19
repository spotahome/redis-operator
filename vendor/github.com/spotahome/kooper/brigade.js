const { events, Job, Group } = require("brigadier")

const githubStatePending = "pending"
const githubStateFailure = "failure"
const githubStateError = "error"
const githubStateSuccess = "success"

const eventTypePush = "push"
const eventTypeExec = "exec"
const eventTypePullRequest = "pull_request"
const eventTypeAfter = "after"
const eventTypeError = "error"
const buildStatusSuccess = "success"
const buildStatusFailure = "failure"
const buildStatusUnhandledRejection = "unhandledRejection"


/**
 * unitTests is the job that will run the unit tests from 
 * the Golang project.
 */
function unitTests() {
    const gopath = "/go"
    const localPath = `${gopath}/src/github.com/spotahome/kooper`

    var job = new Job("unit-test", "golang:1.9");
    job.env = {
        "DEST_PATH": localPath,
        "GOPATH": gopath
    };
    job.tasks = [
        `mkdir -p ${localPath}`,
        `mv /src/* ${localPath}`,
        `cd ${localPath}`,
        "make ci"
    ];

    return job
}


/**
 * fullBuild is the most complete kind of build on the pipeline, 
 * it runs the unit tests.
 */
function fullBuild(e, project) {
    unitTestsJob = unitTests()
    unitTestsJob.run()
}

/**
 * debugFullBuild is like fullbuild but prints the received event information.
 * useful while developing. WARNING: project is not printed because it has secrets
 * and could be leaked by accident.
 */
function debugFullBuild(e, project){
    console.log("-----------------Event-----------------")
    console.log(e)
    console.log("---------------------------------------")
    fullBuild(e, project)
}

/**
 * setGithubCommitStatus returns a job that setups the wanted state on github.
 */
function setGithubCommitStatus(e, project, status) {
    var job = new Job("set-github-build-status", "technosophos/github-notify:latest")
    job.env = {
        GH_REPO: project.repo.name,
        GH_STATE: status,
        GH_DESCRIPTION: `Brigade build finished with ${e.cause.trigger} state`,
        GH_CONTEXT: "brigade",
        GH_TOKEN: project.repo.token,
        GH_COMMIT: e.commit
    }
    return job
}

/**
 * buildStatusToGithub will put the correct build status on github.
 */
function buildStatusToGithub(e, project){
    // Only set state of build when is a push or PR.
    if ([eventTypePush, eventTypePullRequest].includes(e.cause.event.type)) {
        // Set correct status
        var state = githubStateFailure
        if (e.cause.trigger == buildStatusSuccess) {
            state = githubStateSuccess
        }
        // Set the status on github.
        updateGHPRStateJob = setGithubCommitStatus(e, project, state)
        updateGHPRStateJob.run()
    } else {
        console.log(`Build finished with ${e.cause.trigger} state`)
    }
}


events.on(eventTypePush, fullBuild)
events.on(eventTypeExec, debugFullBuild)
events.on(eventTypePullRequest, fullBuild)

// Final events after build (failure or success).
events.on(eventTypeAfter, buildStatusToGithub)
events.on(eventTypeError, buildStatusToGithub)
