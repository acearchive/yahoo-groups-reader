function localizeDateTimes() {
    const allTimeElements = document.querySelectorAll("time[datetime]");
    for (const element of allTimeElements) {
        element.innerHTML = new Intl.DateTimeFormat([], {
            dateStyle: "medium",
            timeStyle: "short",
        }).format(
            new Date(element.getAttribute("datetime"))
        );
    }
}

localizeDateTimes();