function get(field) {
  if (document.forms[0][field].value) {
    return document.forms[0][field].value;
  }
}
function getN(field) {
  if (document.forms[0][field].value) {
    return Number(document.forms[0][field].value);
  }
}
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
function go(url, name, getters) {
  var q = {};

  for (var name in getters) {
    var val = getters[name]();
    if (val) {
      set(name.split('.'), val, q);
    }
  }
  console.log(JSON.stringify(q));
  window.open(url + '/' + name + '?q=' + JSON.stringify(q));
}
function buildEl(prefix, name, def, getters) {
  var prefixedName = prefix ? prefix + '.' + name : name;
  var el = null;
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

  if (el) {
    if (prefixedName == '') {
      prefixedName = 'request';
    }
    var wrap = document.createElement('li');
    var title = document.createElement('label');
    title.innerHTML = prefixedName;
    wrap.appendChild(title);
    wrap.appendChild(el);
    return wrap;
  }
  return null;
}

function build(prefix, name, def) {
  var form = document.getElementById('form');
  var getters = {};

  var el = buildEl(prefix, name, def, getters)
  if (!el) {
    return;
  }
  form.appendChild(el);

  var bt = document.getElementById('go');
  bt.onclick = function() {
    go(urlPath, methodName, getters);
  };
}
