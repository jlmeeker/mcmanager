// Server Actions
function addOp(serverID) {
  var opname = prompt("Player name of the new Op:");
  if (opname != "") {
    var data = new FormData();
    data.append("opname", opname);
    serverAction(serverID, "op/add", data);
  }
}

function backupServer(id) {
  serverAction(id, "backup");
}

function deleteServer(name, id) {
  var r = confirm("Delete "+name+"?");
  if ( r === false) {
    return false;
  }
  serverAction(id, "delete");
}

function setDaytime(id) {
  serverAction(id, "day");
}

function startServer(id) {
  serverAction(id, "start");
}

function stopServer(id) {
  serverAction(id, "stop");
}

function weatherClear(id) {
  serverAction(id, "clear");
}

function whitelistAdd(serverID) {
  var playername = prompt("Name of the player to whitelist:");
  if (playername != "") {
    var data = new FormData();
    data.append("playername", playername);
    serverAction(serverID, "whitelist/add", data);
  }
}

function serverAction(id, action, formdata) {
  var xhttp = new XMLHttpRequest();
  xhttp.onreadystatechange = function() {
    if (this.readyState == 4) {
      var replyObj = JSON.parse(this.responseText);
      if (this.status == 200) {
        document.getElementById('successToastBody').innerText = "Action successful";
        toastList[0].show(); // successToast
      } else {
        document.getElementById('dangerToastBody').innerText = "Error: "+replyObj.error;
        toastList[1].show(); // dangerToast
      }
      fetchServers();
    }
  };
  xhttp.open("POST", "/v1/"+action+"/"+id, true);

  if (formdata instanceof Object) {
    xhttp.send(formdata);
  } else {
    xhttp.send();
  }
}

// Modal Actions
function closeModal(id) {
  var myModalEl = document.getElementById(id);
  var modal = bootstrap.Modal.getInstance(myModalEl)
  modal.hide();
}

// All Forms (new server, login etc.)
function submitForm(loc, form){
  if (loc == "/v1/create" && form.flavor.value == "spigot") {
    document.getElementById('warningToastBody').innerText = "Could take a while, may need to build release.";
    toastList[3]._config.delay = 2000;
    toastList[3].show(); // warningToast
  }

  var data = new FormData(form);
  var xhttp = new XMLHttpRequest();
  xhttp.onreadystatechange = function() {
    if (this.readyState == 4) {
      var replyObj = JSON.parse(this.responseText);
      if (this.status == 200) {
        form.reset();

        document.getElementById('successToastBody').innerText = "Success";
        toastList[0].show(); // successToast

        if (loc == "/v1/create") {
          toastList[3]._config.delay = 5000;
          closeModal('newServerModal');
          if (replyObj.page == "servers") {
            fetchServers();
          } else {
            document.location.href = "/view/servers";
          }
        } else if (loc == "/v1/login") {
          document.getElementById('newServerIcon').classList.remove("hidden");
          document.getElementById('logOutButton').classList.remove("hidden");
          document.getElementById('logInButton').classList.add("hidden");
          document.getElementById('playerName').innerText = replyObj.playername;
          closeModal('logInModal');
        }
      } else {
        document.getElementById('dangerToastBody').innerText = "Error: "+replyObj.error;
        toastList[1].show(); // dangerToast
      }
    }
  };
  xhttp.open("POST",loc, true);
  xhttp.send(data);
  return false;
}

// Logout
function logout() {
  var xhttp = new XMLHttpRequest();
  xhttp.onreadystatechange = function() {
    if (this.readyState == 4) {
      if (this.status == 200) {
        document.getElementById('newServerIcon').classList.add("hidden");
        document.getElementById('logOutButton').classList.add("hidden");
        document.getElementById('logInButton').classList.remove("hidden");
        document.getElementById('playerName').innerText = "";
        document.getElementById('successToastBody').innerText = "Successfully logged out";
        toastList[0].show(); // successToast
        refreshServers({});
      } else {
        document.getElementById('dangerToastBody').innerText = "Error: "+this.responseText;
        toastList[1].show(); // dangerToast
      }
    }
  };
  xhttp.open("POST", "/v1/logout", true);
  xhttp.send();
  return false;
}

// News
function fetchNews() {
  var xhttp = new XMLHttpRequest();
  xhttp.onreadystatechange = function() {
    if (this.readyState == 4) {
      if (this.status == 200) {
        refreshNews(JSON.parse(this.responseText));
      } else {
          document.getElementById('dangerToastBody').innerText = "Error getting news";
          toastList[1].show(); // dangerToast
      }
    }
  };
  xhttp.open("GET", "/v1/news", true);
  xhttp.send(); 
}

function refreshNews(data) {
  var entries = Object.entries(data);
  if (entries.length > 0) {
    document.getElementById("news").innerHTML = "";
    for (const item of entries) {
      newNewsCard(item[1]);
    }
  }
}

function newNewsCard(item) {
  var card = document.createElement("div");
  card.classList.add("col-sm-6", "col-lg-4", "mb-4", "newsitem");
  card.innerHTML = `
  <a href="`+item.preview.Link+`" target="_blank">
    <div class="card shadow newsitem">
      <img class="card-img-top" src="`+item.preview.Images[0]+`">
      <div class="card-body bg-light">
        <h5 class="card-title">`+item.preview.Title+`</h5>
        <p class="card-text">`+item.preview.Description+`</p>
        <p class="card-text">
          <small class="text-muted">`+item.posted+` (`+item.preview.Name+`)</small>
        </p>
      </div>
    </div>
  </a>
  `;
  itemDate = new Date(item.posted);
  if (isToday(itemDate)) {
    card.firstElementChild.classList.add("today");
  }
  document.getElementById("news").appendChild(card);
}

function isToday(d) {
	const today = new Date();
	return d.getDate() == today.getDate() &&
	  d.getMonth() == today.getMonth() &&
	  d.getFullYear() == today.getFullYear()
}

// Servers
function fetchServers() {
  var xhttp = new XMLHttpRequest();
  xhttp.onreadystatechange = function() {
    if (this.readyState == 4) {
      if (this.status == 200) {
        refreshServers(JSON.parse(this.responseText));
      } else {
          document.getElementById('dangerToastBody').innerText = "Error getting servers";
          toastList[1].show(); // dangerToast
      }
    }
  };
  xhttp.open("GET", "/v1/servers", true);
  xhttp.send(); 
}

function refreshServers(data) {
  var entries = Object.entries(data);
  if (entries.length > 0) {
    document.getElementById("servers").innerHTML = "";
    for (var i = 0; i < entries.length; i++) {
      newServerCard(entries[i][1]);
    }
  } else {
    document.getElementById("servers").innerHTML = `
    <div class="text-center lead text-muted">
        <p>Wow, looks pretty empty here...</p>
        <br />
        <p>Time to get your first server up and running.<br />
            Click the <i class="bi bi-minecart-loaded text-success"></i> button at the top of the page.</p>
    </div>
    `;
  }
}

function newServerCard(item) {
  var card = document.createElement("div");
  card.classList.add("col-sm-6", "col-lg-6", "mb-4");
  card.id = item.uuid;
  card.innerHTML = `
    <div class="card shadow">
        <h4 class="card-header bg-light shadow">`+item.name+`
          <div class="mb-0" style="float: right;">
            <a id="backupIndicator_`+item.uuid+`" title="backup" href="#" class="hidden" onClick="backupServer('`+item.uuid+`')">
              <i class="bi-filter-square text-success"></i>
            </a>
            <a id="startIndicator_`+item.uuid+`" title="start" href="#" class="hidden" onClick="startServer('`+item.uuid+`')">
              <i class="bi-caret-right-square text-success"></i>
            </a>
            <a id="stopIndicator_`+item.uuid+`" title="stop" href="#" class="hidden" onClick="stopServer('`+item.uuid+`')">
              <i class="bi-exclamation-square text-warning"></i>
            </a>
            <a id="deleteIndicator_`+item.uuid+`" title="delete" href="#" onclick="deleteServer('`+item.name+`', '`+item.uuid+`')" class="hidden">
              <i class="bi-x-square text-danger"></i>
            </a>
          </div>
        </h4>
        <h4 class="card-header">
            <div class="mb-0">
              <a id="whitelistPlayerIndicator_`+item.uuid+`" title="whitelist player" href="#" class="hidden" onClick="whitelistAdd('`+item.uuid+`')">
                <i class="bi-person-plus text-info"></i>
              </a>
              <a id="addOpIndicator_`+item.uuid+`" title="add op" href="#" class="hidden" onClick="addOp('`+item.uuid+`')">
                <i class="bi-person-lines-fill text-info"></i>
              </a>
              <a id="weatherIndicator_`+item.uuid+`" title="clear weather" href="#" class="hidden" onClick="weatherClear('`+item.uuid+`')">
                <i class="bi-cloud-sun text-primary"></i>
              </a>
              <a id="daytimeIndicator_`+item.uuid+`" title="make daytime" href="#" class="hidden" onClick="setDaytime('`+item.uuid+`')">
                <i class="bi-sunrise text-warning"></i>
              </a>
            </div>
        </h4>
        <div class="card-body bg-light servercard">
            <h6 class="card-title">`+item.motd+`</h6><br>
            <p class="card-text">
              <strong>Flavor:</strong> `+item.flavor+`<br>
              <strong>Release:</strong> `+item.release+`<br>
              <strong>Whitelist:</strong> `+item.whitelistenabled+`<br>
              <strong>Port:</strong> `+item.port+`<br>
              <strong>Autostart:</strong> `+item.autostart+`<br>
              <strong>Ops:</strong> `+item.ops+`<br>
              <strong>Whitelisted:</strong> `+item.whitelist+`<br>
              <strong>Online:</strong> `+item.players+`<br>
            </p>
        </div>
      </div>
    </div>
  `;
  document.getElementById("servers").appendChild(card);
  if (item.running === true) {
    document.getElementById("whitelistPlayerIndicator_"+item.uuid).classList.remove("hidden");
    document.getElementById("addOpIndicator_"+item.uuid).classList.remove("hidden");
    document.getElementById("weatherIndicator_"+item.uuid).classList.remove("hidden");
    document.getElementById("daytimeIndicator_"+item.uuid).classList.remove("hidden");
    document.getElementById("backupIndicator_"+item.uuid).classList.remove("hidden");
    document.getElementById("stopIndicator_"+item.uuid).classList.remove("hidden");
  } else {
    document.getElementById("startIndicator_"+item.uuid).classList.remove("hidden");
  }
  if (item.amowner === true) {
    document.getElementById("deleteIndicator_"+item.uuid).classList.remove("hidden");
  }
}