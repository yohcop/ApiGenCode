// Set the given value at the given path, in obj.
function set(path, value, obj) {
  var name = path[0];

  var isArray = false;
  var m = name.match(/^\[(\d+)\]$/)
  if (m) {
    name = Number(m[1])
  } else {
    var arrayName = name.match(/(.+)\[\]/);
    if (arrayName) {
      name = arrayName[1];
      isArray = true;
    }
  }

  if (path.length == 1) {
    obj[name] = value;
  } else if (name in obj) {
    set(path.slice(1), value, obj[name]);
  } else {
    var s = null;
    if (isArray) {
      s = [];
    } else {
      s = {}
    }
    set(path.slice(1), value, s);
    obj[name] = s;
  }
}
// Prepare the request and open the new url in a new window.
function go(url, path, getters) {
  var q = {};

  for (var field in getters) {
    var val = getters[field]();
    if (val) {
      set(field.split('.'), val, q);
    }
  }
  console.log(JSON.stringify(q));
  window.open(url + path + '?q=' + JSON.stringify(q));
}
// Build a <li> tag that contains the form fields for the given definition.
// Sets callback in getters for the defined fields.
function buildEl(prefix, name, def, getters) {
  var prefixedName = prefix ? prefix + '.' + name : name;
  var el = null;

  if (!('enum' in def)) {
    switch(def.type) {
      case "string":
        el = document.createElement('input');
        //el.name = prefixedName;
        getters[prefixedName] = (function() { return function() {
            if (el.value) {
              return el.value;
            }
            return null;
        }})();
        break;
      case "number":
        el = document.createElement('input');
        //el.name = prefixedName;
        getters[prefixedName] = (function() { return function() {
            if (el.value) {
              return Number(el.value);
            }
            return null;
        }})();
        break;
      case "object":
        el = document.createElement('ul');
        for (var sub in def.properties) {
          var child = buildEl(
              prefixedName, sub, def.properties[sub], getters);
          if (child) {
            el.appendChild(child);
          }
        }
        break;
      case "array":
        var items = def.items;
        el = document.createElement('div');
        var children = document.createElement('ul');
        el.appendChild(children);
        var bt = document.createElement('button');
        bt.innerHTML = 'Add';
        el.appendChild(bt);

        var count = 0;
        bt.onclick = (function() {
          return function() {
            var child = buildEl(
              prefixedName + '[]', '[' + count + ']', items, getters);
            children.appendChild(child);
            count++;
            return false;
          }
        })();
        break;
      default:
        if (def.$ref) {
          var d = schemas[def.$ref];
          if (d) {
            return buildEl(prefix, name, d, getters);
          }
        }
    }
  } else {
    // Enums are handled separately since they also may have
    // a 'type' attribute, but we just ignore it and take the
    // values in the enum array.
    el = document.createElement('select');
    for (var i = 0; i < def.enum.length; ++i) {
      var opt = document.createElement('option');
      opt.value = i;
      opt.innerHTML = JSON.stringify(def.enum[i]);
      el.appendChild(opt);
    }
    getters[prefixedName] = (function() { return function() {
        if (el.value) {
          return def.enum[el.value];
        }
        return null;
    }})();
  }

  if (el) {
    prefixedName = prefixedName || 'request';
    var wrap = document.createElement('li');
    var title = document.createElement('label');
    title.innerHTML = prefixedName;
    wrap.appendChild(title);
    wrap.appendChild(el);
    return wrap;
  }
  return null;
}
// Prepare the form.
function build(prefix, name, def) {
  var form = document.getElementById('form');
  var getters = {};

  var el = buildEl(prefix, name, def, getters)
  if (!el) {
    return;
  }
  form.appendChild(el);

  var bt = document.getElementById('go');
  var p = methodPath || ('/' + methodName);
  bt.onclick = function() {
    go(urlPath, p, getters);
  };
}
