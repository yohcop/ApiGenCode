// These are globals, since setup() can be called multiple times,
// and we want to share messages and methods.
var messages = {};
var enums = {};
var methods = {};

function updateMethods() {
  var list = document.getElementById('methods');
  for (var m in methods) {
    var el = document.createElement('option');
    el.innerHTML = m;
    el.value = m;
    list.appendChild(el);
  }
}

function selectmethod() {
  setupForMethod(
      document.getElementById('methods').value);
}

function buildFormForField(field, path) {
  switch (field.label) {
    case 'LABEL_REQUIRED':
      return buildFormForReqField(field, path);
    case 'LABEL_OPTIONAL':
      return buildFormForOptField(field, path);
    case 'LABEL_REPEATED':
      return buildFormForRepeatedField(field, path);
  }
  return null;
}

function buildFormForReqField(field, path) {
  var el = null;
  var getter = null;

  switch (field.type) {
    case "TYPE_STRING":
      el = document.createElement('input');
      getter = (function() { return function() {
        if (el.value) return el.value;
        return "";
      }})();
      break;
    case "TYPE_INT32":
      el = document.createElement('input');
      getter = (function() { return function() {
        if (el.value) return Math.floor(Number(el.value));
        return 0;
      }})();
      break;
    case "TYPE_FLOAT":
      el = document.createElement('input');
      getter = (function() { return function() {
        if (el.value) return Number(el.value);
        return 0;
      }})();
      break;
    case "TYPE_BOOL":
      el = document.createElement('input');
      el.type = 'checkbox';
      getter = (function() { return function() {
        return el.checked;
      }})();
      break;
    case "TYPE_ENUM":
      var def = enums[field.type_name];
      el = document.createElement('select');
      for (var i = 0; i < def.value.length; ++i) {
        var e = def.value[i];
        var opt = document.createElement('option');
        opt.value = i;
        opt.innerHTML = e.name;
        el.appendChild(opt);
      }
      getter = (function() { return function() {
        if (el.value) {
          return def.value[el.value].number;
        }
        return null;
      }})();
      break;
    case "TYPE_MESSAGE":
      var def = messages[field.type_name];
      var req = buildFormForMessage(def, path);
      getter = (function() { return function() {
        return req.getter();
      }})();
      el = req.el;
      break;
    case "TYPE_BYTES":
      return null;
    default:
      console.log('Unhandled: ', field);
  }

  if (el) {
    var wrap = document.createElement('li');
    var title = document.createElement('label');
    title.innerHTML = field.name;
    wrap.appendChild(title);
    wrap.appendChild(el);
    return {el:wrap, getter:getter};
  }
  return null;
}

function buildFormForOptField(field, path) {
  var req = buildFormForReqField(field, path)
  if (!req) {
    return null;
  }

  var opt = document.createElement('input');
  opt.type='checkbox';
  req.el.insertBefore(opt, req.el.firstChild);

  var getter = (function() { return function() {
    var v = req.getter();
    if (!opt.checked) return null;
    return req.getter();
  }})();

  return {el:req.el, getter:getter};
}

function buildFormForRepeatedField(field, path) {
  var el = document.createElement('li');
  var name = document.createElement('span');
  var children = document.createElement('ul');
  var add = document.createElement('button');
  name.innerHTML = field.name;
  add.innerHTML = 'Add';
  el.appendChild(name);
  el.appendChild(children);
  el.appendChild(add);

  var childrenArray = [];
  add.onclick = (function() { return function() {
    var req = buildFormForReqField(field, path)
    if (!req) {
      return null;
    }
    var rem = document.createElement('button');
    rem.innerHTML = 'Remove';
    req.el.appendChild(rem);
    children.appendChild(req.el);
    childrenArray.push(req);

    rem.onclick = (function() { return function() {
      req.el.parentNode.removeChild(req.el);
      var at = childrenArray.indexOf(req);
      childrenArray.splice(at, 1);
    }})();
  }})();

  var getter = (function() { return function() {
    var res = [];
    for (var i = 0; i < childrenArray.length; i++) {
      res.push(childrenArray[i].getter());
    }
    if (res.length) {
      return res;
    }
    return null;
  }})();

  return {el:el, getter:getter};
}

function buildFormForMessage(message, path) {
  var el = document.createElement('ul');

  var getters = {};
  for (var i in message.field) {
    var f = message.field[i];
    var c = buildFormForField(f, path);
    if (!c) {
      continue;
    }
    el.appendChild(c.el);
    getters[message.field[i].name] = c.getter;
  }
  var getter = (function() { return function() {
    var res = {};
    for (var name in getters) {
      var v = getters[name]();
      if (v != null && v != undefined) {
        res[name] = v;
      }
    }
    return res;
  }})();
  return {el:el, getter:getter};
}

function setupForMethod(methodName) {
  var m = methods[methodName];
  if (!m) {
    return;
  }

  var form = document.getElementById('form');
  var go = document.getElementById('go');
  var jsonOutput = document.getElementById('jsonreq');
  var reqFrameUrl = document.getElementById('reqframeurl');
  var reqFrame = document.getElementById('reqframe');

  while (form.hasChildNodes()) {
    form.removeChild(form.lastChild);
  }

  var input = m.input_type;
  req = buildFormForMessage(messages[input], input);
  form.appendChild(req.el);
  go.onclick = function() {
    var obj = req.getter();
    jsonOutput.innerHTML = JSON.stringify(obj, "", "  ");
    var url = m.url + '?q=' + JSON.stringify(obj);
    reqFrameUrl.innerHTML = url;
    reqFrame.src = url;
  };
}

function setup(descriptor) {
  var pkg = descriptor.package;

  // Messages
  for (var i in descriptor.message_type) {
    var m = descriptor.message_type[i];
    messages['.' + pkg + '.' + m.name] = m;
  }

  // Enums
  for (var i in descriptor.enum_type) {
    var m = descriptor.enum_type[i];
    enums['.' + pkg + '.' + m.name] = m;
  }

  // Services
  for (var i in descriptor.service) {
    var s = descriptor.service[i];
    var n = s.name;
    for (var j in s.method) {
      var m = s.method[j];
      methods['.' + pkg + '.' + n + '.' + m.name] = m;
    }
  }

  updateMethods();
  selectmethod();
}

function setServiceUrl(service, url) {
  methods[service].url = url;
}
