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
