function HookZippyEvents() {
	for (var el of document.querySelectorAll("details")) {
		if (!el.id.startsWith("moves")) {
			continue;
		}
		el.addEventListener("toggle", SaveZippies)
	}
}


function LoadZippies() {
	for (var el of document.querySelectorAll("details")) {
		if (!el.id.startsWith("moves")) {
			continue;
		}
		var data = sessionStorage.getItem(el.id);

		if (data) {
			el.open = true;
		}
	}

}

function SaveZippies() {
	for (var el of document.querySelectorAll("details")) {
		if (!el.id.startsWith("moves")) {
			continue;
		}
		if (el.open) {
			sessionStorage.setItem(el.id, true);
		} else {
			sessionStorage.removeItem(el.id);
		}
	}
}

function SetUp() {
	LoadZippies();
	HookZippyEvents();
}

SetUp();