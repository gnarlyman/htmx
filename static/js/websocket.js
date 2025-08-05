const ws = new WebSocket("ws://" + window.location.host + "/ws");
ws.onmessage = function (event) {
    if (event.data === "new-room") {
        htmx.trigger("#rooms-list", "new-room");
    } else if (event.data === "new-chat") {
        htmx.trigger("#messages-list", "new-chat");
    }
};
ws.onclose = function () {
    setTimeout(() => location.reload(), 1000);
    // console.error("WebSocket connection closed. Reconnecting...");
};