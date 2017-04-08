"use strict";

var query = {};
var meta = {};

var metric_prefixes = [
    "µs", "ms", "s"
];

function set(param, val) {
    var el = document.forms.item(0)[param];
    if (el && el.dataset && el.dataset.default) {
        if (val == el.dataset.default) {
            delete query[param];
        } else {
            query[param] = val;
        }
    } else {
        query[param] = val;
    }
    localStorage["query"] = JSON.stringify(query);

    update();
}

function setShutterSpeed() {
    var form = document.forms.item(0);
    var ratio = form["ss-ratio"].value;
    var val = form["ss"].value;
    var ss = ratio * val;

    if (ss === 0) {
        delete query["ss"];
        delete meta["ss"];
        delete meta["ss-ratio"];
    } else {
        query["ss"] = ss;
        meta["ss"] = val;
        meta["ss-ratio"] = ratio;
    }
    localStorage["query"] = JSON.stringify(query);
    localStorage["meta"] = JSON.stringify(meta);

    update();
}

function create_params() {
    let params = Object.keys(query).map((val) => {
        return val + "=" + query[val];
    });
    return params.join("&");
}

function updatePreview(param, val) {
    var previewEl = document.getElementById(param + "-preview");
    if (previewEl) {
        previewEl.innerHTML = val;
    }
}

function updateShutterSpeed() {
    var previewEl = document.getElementById("ss-preview");
    var unitEl = document.getElementById("ss-preview-unit");
    var form = document.forms.item(0);

    var ratio = form["ss-ratio"].value;
    var val = form["ss"].value;
    var ss = ratio * val;

    if (ss === 0) {
        previewEl.innerHTML = "auto";
        unitEl.style.display = "none";
    } else {
        var step = 0;
        while (ss > 1000) {
            ss /= 1000;
            step++;
        }
        previewEl.innerHTML = ss;
        unitEl.innerHTML = metric_prefixes[step];
        unitEl.style.display = "";
    }
}

function update() {
    var full = document.getElementById("full");
    var preview = document.getElementById("preview");
    var params = create_params();
    preview.src = "preview?" + params;
    full.href = "full?" + params;
}

function set_default(id) {
    var el = document.forms.item(0)[id];

    var value;
    if (query[id]) {
        value = query[id];
    } else if (el.dataset && el.dataset.default) {
        value = el.dataset.default;
    }

    switch (el.type) {
        case "checkbox":
        case "radio":
            el.checked = (String(value) == "true") ? true : false;
            break;
        default:
            el.value = value;
            break;
    }

    updatePreview(id, value);
}

function set_defaultShutterSpeed() {
    var form = document.forms.item(0);
    var ratio = form["ss-ratio"];
    var val = form["ss"];

    if (meta["ss"]) {
        val.value = meta["ss"];
    }
    if (meta["ss-ratio"]) {
        ratio.value = meta["ss-ratio"];
    }
}

document.addEventListener("DOMContentLoaded", function(event) {
    if (localStorage["query"]) {
        query = JSON.parse(localStorage["query"]);
    } else {
        localStorage["query"] = "";
    }

    if (localStorage["meta"]) {
        meta = JSON.parse(localStorage["meta"]);
    } else {
        localStorage["meta"] = "";
    }

    set_default("exposure");
    set_default("awb");
    set_default("ifx");
    set_default("iso");
    set_default("ss");
    set_default("hf");
    set_default("vf");
    set_defaultShutterSpeed();
    updateShutterSpeed();
    update();
});
