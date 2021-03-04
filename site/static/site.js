function servers() {
   fetch('/v1/servers')
  .then(response => response.json())
  .then(data => console.log(data));
}

function startServer(id) {
  serverStartStop(id, "start");
}

function stopServer(id) {
  serverStartStop(id, "stop");
}

function serverStartStop(id, action) {
  var xhttp = new XMLHttpRequest();
  xhttp.onreadystatechange = function() {
    if (this.readyState == 4) {
      console.log(this.responseText);
    }
  };
  xhttp.open("POST", "/v1/"+action+"/"+id, true);
  xhttp.send();
}

function deleteServer(id) {
  var xhttp = new XMLHttpRequest();
  xhttp.onreadystatechange = function() {
    if (this.readyState == 4) {
      console.log(this.responseText);
      if (this.status == 200) {
        location.reload();
      }
    }
  };
  xhttp.open("POST", "/v1/delete/"+id, true);
  xhttp.send();
}

function fetchServerStatuses() {
  var xhttp = new XMLHttpRequest();
  xhttp.onreadystatechange = function() {
    if (this.readyState == 4 && this.status == 200) {
      refreshServerCards(JSON.parse(this.responseText));
    }
  };
  xhttp.open("GET", "/v1/servers", true);
  xhttp.send(); 
}

function refreshServerCards(data) {
  for (const [name, status] of Object.entries(data)) {
    refreshCard(status);
  }
}

function refreshCard(status) {
  try {
    var card = document.getElementById("card_"+status.uuid);
    var junk = card.innerText;
  } 
  catch(err) {
    console.log(err);
    return;
  }
  document.getElementById("port_"+status.uuid).innerText = status.port;
  document.getElementById("autostart_"+status.uuid).innerText = status.autostart;
  document.getElementById("players_"+status.uuid).innerText = status.players;
  document.getElementById("flavor_"+status.uuid).innerText = status.flavor;
  document.getElementById("ops_"+status.uuid).innerText = status.ops;

  var btn = document.getElementById("btn_"+status.uuid);
  if (status.running) {
    btn.classList.remove("btn-success");
    btn.classList.add("btn-warning");
    btn.onclick = function() { stopServer(status.uuid); };
    btn.innerText = "Stop";
  } else {
    btn.classList.remove("btn-warning");
    btn.classList.add("btn-success");
    btn.onclick = function() { startServer(status.uuid); };
    btn.innerText = "Start";
  }
}

function submitForm(loc, form){
  var data = new FormData(form);
  var xhttp = new XMLHttpRequest();
  xhttp.onreadystatechange = function() {
    if (this.readyState == 4 && this.status == 200) {
      form.reset();
      if (loc == "/v1/create") {
        console.log(this.responseText);
      } 
      location.reload();
    }
  };
  xhttp.open("POST",loc, true);
  xhttp.send(data);
  return false;
}

function logout() {
  var xhttp = new XMLHttpRequest();
  xhttp.onreadystatechange = function() {
    if (this.readyState == 4) {
      location.reload();
    }
  };
  xhttp.open("POST", "/v1/logout", true);
  xhttp.send();
  return false;
}