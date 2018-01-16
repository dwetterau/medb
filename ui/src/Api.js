import $ from 'jquery';
import {notify} from 'react-notify-toast';

export function handlePull() {
    $.get("/api/1/pull", () => {
        notify.show('Pulled successfully.')
    })
}

export function handlePush() {
    $.get("/api/1/push", () => {
        notify.show('Pushed successfully.')
    })
}

export function fetchGitInfo(callback) {
    $.get("/api/1/git/info", (info) => {
        callback(JSON.parse(info))
    })
}

export function emptyGitInfo() {
    return {
        lastCommit: "",
        lastPull: "",
        remoteAheadBy: "",
        localAheadBy: "",
    }
}
