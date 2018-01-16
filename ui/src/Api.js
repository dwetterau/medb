import $ from 'jquery';

export function handlePull() {
    $.get("/api/1/pull", () => {
        window.location.reload()
    })
}

export function handlePush() {
    $.get("/api/1/push", () => {
        window.location.reload()
    })
}
