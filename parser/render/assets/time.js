function hasTimeComponent(date) {
    return date.getUTCHours() !== 0
        || date.getUTCMinutes() !== 0
        || date.getUTCSeconds() !== 0
        || date.getUTCMilliseconds() !== 0
}

function localizeDateTimes() {
    const allTimeElements = document.querySelectorAll("time[datetime]");
    for (const element of allTimeElements) {
        const timestamp = new Date(element.getAttribute("datetime"))
        const options = hasTimeComponent(timestamp) ? {
            dateStyle: "medium",
            timeStyle: "short",
        } : {
            dateStyle: "medium",
        }

        element.innerHTML = new Intl.DateTimeFormat([], options).format(timestamp);
    }
}

localizeDateTimes();