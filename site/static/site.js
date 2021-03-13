// Server Actions
function addOp(serverID) {
  var opname = prompt("Player name of the new Op:");
  if (opname != "" && opname != null) {
    var data = new FormData();
    data.append("opname", opname);
    serverAction(serverID, "op/add", data);
  }
}

function backupServer(id) {
  serverAction(id, "backup");
}


function saveServer(id) {
  serverAction(id, "save");
}

function deleteServer(name, id) {
  var r = confirm("Delete " + name + "?\n\nTHIS CANNOT BE UNDONE !!!");
  if (r === false) {
    return false;
  }
  serverAction(id, "delete");
}

function regenServer(name, id) {
  var r = confirm("Regen " + name + "?\n\nTHIS WILL DELETE ALL WORLD AND IN-GAME PLAYER DATA !!!");
  if (r === false) {
    return false;
  }
  serverAction(id, "regen");
}

function setDaytime(id) {
  serverAction(id, "time/day");
}

function startServer(id) {
  serverAction(id, "start");
}

function stopServer(id) {
  serverAction(id, "stop");
}

function weatherClear(id) {
  serverAction(id, "weather/clear");
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
  xhttp.onreadystatechange = function () {
    if (this.readyState == 4) {
      var replyObj = JSON.parse(this.responseText);
      if (this.status == 200) {
        document.getElementById('successToastBody').innerText = "Action successful";
        toastList[0].show(); // successToast

        if (action == "delete") {
          document.getElementById("servers").removeChild(document.getElementById(id));
        }
      } else {
        document.getElementById('dangerToastBody').innerText = "Error: " + replyObj.error;
        toastList[1].show(); // dangerToast
      }
      fetchServers();
    }
  };
  xhttp.open("POST", "/api/v1/" + action + "/" + id, true);

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
function submitForm(loc, form) {
  if (loc == "/api/v1/create" && form.flavor.value == "spigot") {
    document.getElementById('warningToastBody').innerText = "Could take a while, may need to build release.";
    toastList[3]._config.delay = 2000;
    toastList[3].show(); // warningToast
  }

  var data = new FormData(form);
  var xhttp = new XMLHttpRequest();
  xhttp.onreadystatechange = function () {
    if (this.readyState == 4) {
      var replyObj = JSON.parse(this.responseText);
      if (this.status == 200) {
        form.reset();
        document.getElementById('successToastBody').innerText = "Success";
        toastList[0].show(); // successToast

        if (loc == "/api/v1/create") {
          toastList[3]._config.delay = 5000;
          closeModal('newServerModal');
          if (replyObj.page == "servers") {
            fetchServers();
          } else {
            document.location.href = "/view/servers";
          }
        } else if (loc == "/api/v1/login") {
          document.getElementById('newServerIcon').classList.remove("hidden");
          document.getElementById('logOutButton').classList.remove("hidden");
          document.getElementById('logInButton').classList.add("hidden");
          document.getElementById('playerName').innerText = replyObj.playername;
          closeModal('logInModal');
          if (replyObj.page == "servers") {
            fetchServers();
          }
        }
      } else {
        document.getElementById('dangerToastBody').innerText = "Error: " + replyObj.error;
        toastList[1].show(); // dangerToast
      }
    }
  };
  xhttp.open("POST", loc, true);
  xhttp.send(data);
  return false;
}

// Logout
function logout() {
  var xhttp = new XMLHttpRequest();
  xhttp.onreadystatechange = function () {
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
        document.getElementById('dangerToastBody').innerText = "Error: " + this.responseText;
        toastList[1].show(); // dangerToast
      }
    }
  };
  xhttp.open("POST", "/api/v1/logout", true);
  xhttp.send();
  return false;
}

// News
function fetchNews() {
  var xhttp = new XMLHttpRequest();
  xhttp.onreadystatechange = function () {
    if (this.readyState == 4) {
      if (this.status == 200) {
        refreshNews(JSON.parse(this.responseText));
      } else {
        document.getElementById('dangerToastBody').innerText = "Error getting news";
        toastList[1].show(); // dangerToast
      }
    }
  };
  xhttp.open("GET", "/api/v1/news", true);
  xhttp.send();
}

function refreshNews(data) {
  var entries = Object.entries(data.news);
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
  <a href="`+ item.preview.Link + `" target="_blank">
    <div class="card shadow newsitem">
      <img class="card-img-top" src="`+ item.preview.Images[0] + `">
      <div class="card-body bg-light">
        <h5 class="card-title">`+ item.preview.Title + `</h5>
        <p class="card-text">`+ item.preview.Description + `</p>
        <p class="card-text">
          <small class="text-muted">`+ item.posted + ` (` + item.preview.Name + `)</small>
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
  xhttp.onreadystatechange = function () {
    if (this.readyState == 4) {
      if (this.status == 200) {
        refreshServers(JSON.parse(this.responseText));
      } else {
        document.getElementById('dangerToastBody').innerText = "Error getting servers";
        toastList[1].show(); // dangerToast
      }
    }
  };
  xhttp.open("GET", "/api/v1/servers", true);
  xhttp.send();
}

function refreshServers(data) {
  if (data.hasOwnProperty("servers")) {
    var entries = Object.entries(data.servers);
    document.getElementById("noservers").classList.add("hidden");
    for (var i = 0; i < entries.length; i++) {
      refreshServerCard(entries[i][1]);
    }
  } else {
    var scards = document.getElementsByClassName('scard');
    while (scards[0]) {
      scards[0].parentNode.removeChild(scards[0]);
    }
    document.getElementById("noservers").classList.remove("hidden");
  }
}

function newServerCard(item) {
  var card = document.createElement("div");
  card.classList.add("col-sm-6", "col-lg-6", "mb-4");
  card.id = item.uuid;
  card.innerHTML = `
    <div class="card shadow">
        <h4 class="card-header bg-light shadow">`+ item.name + `
          <div class="mb-0" style="float: right;">
            <a id="regenIndicator_`+ item.uuid + `" title="regen world (DESTRUCTIVE)" href="#" class="hidden" onClick="regenServer('` + item.name + `', '` + item.uuid + `')">
              <i class="bi-card-image text-success"></i>
            </a>
            <a id="startIndicator_`+ item.uuid + `" title="start" href="#" class="hidden" onClick="startServer('` + item.uuid + `')">
              <i class="bi-caret-right-square text-success"></i>
            </a>
            <a id="stopIndicator_`+ item.uuid + `" title="stop" href="#" class="hidden" onClick="stopServer('` + item.uuid + `')">
              <i class="bi-exclamation-square text-warning"></i>
            </a>
            <a id="deleteIndicator_`+ item.uuid + `" title="delete (DESTRUCTIVE)" href="#" onclick="deleteServer('` + item.name + `', '` + item.uuid + `')" class="hidden">
              <i class="bi-x-square text-danger"></i>
            </a>
          </div>
        </h4>
        <h4 class="card-header">
          <div class="mb-0">
            <a id="whitelistPlayerIndicator_`+ item.uuid + `" title="whitelist player" href="#" class="hidden" onClick="whitelistAdd('` + item.uuid + `')">
              <i class="bi-person-plus text-success"></i>
            </a>
            <a id="addOpIndicator_`+ item.uuid + `" title="add op" href="#" class="hidden" onClick="addOp('` + item.uuid + `')">
              <i class="bi-person-lines-fill text-info"></i>
            </a>
            <a id="weatherIndicator_`+ item.uuid + `" title="clear weather" href="#" class="hidden" onClick="weatherClear('` + item.uuid + `')">
              <i class="bi-cloud-sun text-primary"></i>
            </a>
            <a id="daytimeIndicator_`+ item.uuid + `" title="make daytime" href="#" class="hidden" onClick="setDaytime('` + item.uuid + `')">
              <i class="bi-sunrise text-warning"></i>
            </a>
            <a id="backupIndicator_`+ item.uuid + `" title="backup" href="#" class="hidden" onClick="backupServer('` + item.uuid + `')">
            <i class="bi-filter-square text-primary"></i>
            </a>
            <a id="saveIndicator_`+ item.uuid + `" title="save" href="#" class="hidden" onClick="saveServer('` + item.uuid + `')">
              <i class="bi-save2 text-success"></i>
            </a>
          </div>
        </h4>
        <div class="card-body bg-light servercard">
          <h6 class="card-title"><strong>Server Address:</strong> <br /><span class="text-success">`+ hostname + ":" + item.port + `</span></h6>
          <p class="card-text">
            <strong>MOTD:</strong> `+ item.motd + `<br>
            <strong>Flavor:</strong> `+ item.flavor + `<br>
            <strong>Release:</strong> `+ item.release + `<br>
            <strong>Game Mode:</strong> `+ item.gamemode + `<br>
            <strong>World Type:</strong> `+ item.worldtype + `<br>
            <strong>Seed:</strong> `+ item.seed + `<br>
            <strong>Whitelist On:</strong> `+ item.whitelistenabled + `<br>
            <strong>Hardcore:</strong> `+ item.hardcore + `<br>
            <strong>PVP:</strong> `+ item.pvp + `<br>
            <strong>Autostart:</strong> `+ item.autostart + `<br>
            <strong>Ops:</strong> `+ item.ops + `<br>
            <strong>Whitelisted:</strong> `+ item.whitelist + `<br>
            <strong>Online:</strong> `+ item.players + `<br>
          </p>
        </div>
      </div>
    </div>
  `;
  document.getElementById("servers").appendChild(card);
  if (item.running === true) {
    document.getElementById("addOpIndicator_" + item.uuid).classList.remove("hidden");
    document.getElementById("weatherIndicator_" + item.uuid).classList.remove("hidden");
    document.getElementById("daytimeIndicator_" + item.uuid).classList.remove("hidden");
    document.getElementById("backupIndicator_" + item.uuid).classList.remove("hidden");
    document.getElementById("saveIndicator_" + item.uuid).classList.remove("hidden");
    document.getElementById("stopIndicator_" + item.uuid).classList.remove("hidden");

    if (item.whitelistenabled === true) {
      document.getElementById("whitelistPlayerIndicator_" + item.uuid).classList.remove("hidden");
    }
  } else {
    document.getElementById("startIndicator_" + item.uuid).classList.remove("hidden");
  }
  if (item.amowner === true) {
    document.getElementById("deleteIndicator_" + item.uuid).classList.remove("hidden");
    document.getElementById("regenIndicator_" + item.uuid).classList.remove("hidden");
  }
}


function newServerCardV2(item) {
  var card = document.createElement("div");
  card.classList.add("scard", "col-sm-6", "col-md-6", "col-lg-4", "mb-4");
  card.id = item.uuid;
  card.innerHTML = `
    <div class="card text-secondary">
      <div class="card-header mx-0">
        <div class="dropdown">
          <a class="btn btn-secondary dropdown-toggle" href="#" role="button" id="dropdownMenuLink`+ item.uuid + `" data-bs-toggle="dropdown" aria-expanded="false">
            <i class="bi-list"></i>
          </a>
          <ul class="dropdown-menu" aria-labelledby="dropdownMenuLink`+ item.uuid + `">
            <li>
              <a id="whitelistPlayerIndicator_`+ item.uuid + `" title="whitelist player" href="#" class="dropdown-item disabled" onClick="whitelistAdd('` + item.uuid + `')">
                <i class="bi-person-plus text-success"></i> Whitelist Player
              </a>
            </li>
            <li>
              <a id="addOpIndicator_`+ item.uuid + `" title="add op" href="#" class="dropdown-item disabled" onClick="addOp('` + item.uuid + `')">
                <i class="bi-person-lines-fill text-info"></i> Add Op
              </a>
            </li>
            <li>
              <a id="weatherIndicator_`+ item.uuid + `" title="clear weather" href="#" class="dropdown-item disabled" onClick="weatherClear('` + item.uuid + `')">
                <i class="bi-cloud-sun text-primary"></i> Weather Clear
              </a>
            </li>
            <li>
              <a id="daytimeIndicator_`+ item.uuid + `" title="make daytime" href="#" class="dropdown-item disabled" onClick="setDaytime('` + item.uuid + `')">
                <i class="bi-sunrise text-warning"></i> Set Daytime
              </a>
            </li>
            <li>
              <a id="backupIndicator_`+ item.uuid + `" title="backup" href="#" class="dropdown-item" onClick="backupServer('` + item.uuid + `')">
                <i class="bi-filter-square text-primary"></i> Backup
              </a>
            </li>
            <li>
              <a id="saveIndicator_`+ item.uuid + `" title="save" href="#" class="dropdown-item disabled" onClick="saveServer('` + item.uuid + `')">
                <i class="bi-save2 text-success"></i> Save
              </a>
            </li>
            <li>
              <a id="deleteIndicator_`+ item.uuid + `" title="save" href="#" class="dropdown-item disabled" onClick="deleteServer('` + item.name + `', '` + item.uuid + `')">
                <i class="bi-trash text-danger"></i> DELETE
              </a>
            </li>
          </ul>
        </div>
      </div>
      <div id="carouselControls_`+ item.uuid + `" class="carousel carousel-dark slide" data-bs-ride="carousel" data-bs-interval="0">
        <div class="carousel-inner">
          <div class="carousel-item active">
          
            <div class="card-body bg-light">
              <div class="card-text servercard">
                <div class="text-center">
                  <div class="serverName">
                    <h1 id="name_`+ item.uuid + `">
                      `+ item.name + `
                    </h1>
                    <span id="address_`+ item.uuid + `" class="text-success">` + hostname + ":" + item.port + `</span> 
                  </div>
                  <h4 class="serverState">
                    <span id="running_`+ item.uuid + `" class="text-success">` + runningToString(item.running) + `</span>
                  </h4>
                  <h4 id="motd_`+ item.uuid + `" class="serverMOTD">` + item.motd + `</h4>
                </div>
              </div>
            </div>
          
          </div>
          <div class="carousel-item">
            <div class="card-body bg-light">
              <div class="servercard">
                <div class="text-center">
                  <button id="regenIndicator_`+ item.uuid + `" type="button" class="btn btn-lg btn-light text-warning border-warning" title="regen world (DESTRUCTIVE)" href="#" onclick="regenServer('` + item.name + `', '` + item.uuid + `')" disabled>
                    <i class="bi-card-image"></i>
                  </button>
                  <br />
                  <br />
                  <button id="startIndicator_`+ item.uuid + `" type="button" class="btn btn-lg btn-light text-success border-success" title="start" href="#" onClick="startServer('` + item.uuid + `')" disabled>
                    <i class="bi-caret-right-square"></i>
                  </button>
                  <br />
                  <br />
                  <button id="stopIndicator_`+ item.uuid + `" type="button" class="btn btn-lg btn-light text-danger border-danger" title="stop" href="#" onClick="stopServer('` + item.uuid + `')" disabled>
                    <i class="bi-exclamation-octagon"></i>
                  </button>
                  </div>
              </div>
            </div>
          </div>
          <div class="carousel-item">
            <div class="card-body bg-light">
              <div class="servercard">
                <div class="text-center">
                  <strong>Online Players:</strong><br />
                  <span id="players_`+ item.uuid + `">` + listToVertical(item.players) + `</span>
                </div>
              </div>
            </div>
          </div>
          <div class="carousel-item">
            <div class="card-body bg-light">
              <div class="card-text servercard">
                <div class="text-center">
                  <p class="">
                    <strong>Game Mode:</strong> `+ item.gamemode + `<br>
                    <strong>World Type:</strong> `+ item.worldtype + `<br>
                    <strong>Seed:</strong> `+ item.seed + `<br>
                    <strong>Whitelist On:</strong> `+ item.whitelistenabled + `<br>
                    <strong>Hardcore:</strong> `+ item.hardcore + `<br>
                    <strong>PVP:</strong> `+ item.pvp + `<br>
                    <strong>Autostart:</strong> `+ item.autostart + `<br>
                    <strong>Ops:</strong> `+ item.ops + `<br>
                    <strong>Whitelisted:</strong> `+ item.whitelist + `<br>
                  </p>
                </div>
              </div>
            </div>
          </div>
          <!-- 
          <div class="carousel-item">
            <div class="card-body bg-light">
              <div class="servercard">
                <div class="text-center">
                </div>
              </div>
            </div>
          </div>
          -->
        </div>
        <button class="carousel-control-prev" type="button" data-bs-target="#carouselControls_`+ item.uuid + `"  data-bs-slide="prev">
          <span class="carousel-control-prev-icon" aria-hidden="true"></span>
          <span class="visually-hidden">Previous</span>
        </button>
        <button class="carousel-control-next" type="button" data-bs-target="#carouselControls_`+ item.uuid + `"  data-bs-slide="next">
          <span class="carousel-control-next-icon" aria-hidden="true"></span>
          <span class="visually-hidden">Next</span>
        </button>
      </div>
      <div class="card-footer">
        <div class="float-end col-4 footer-text">
          <strong>Online</strong><br />
          <span id="online_`+ item.uuid + `">` + countNonEmpty(item.players) + `</span>
        </div>
        <div class="float-end col-4 footer-text">
          <strong>Flavor</strong><br />
          <span id="flavor_`+ item.uuid + `">` + item.flavor + `</span>
        </div>
        <div class="float-end col-4 footer-text">
          <strong>Release</strong><br />
          <span id="release_`+ item.uuid + `">` + item.release + `</span>
        </div>
      </div>
    </div>
  `;
  document.getElementById("servers").appendChild(card);
  updateCardActionButtons(item);
}

function refreshServerCard(serverData) {
  var svr = document.getElementById(serverData.uuid);
  if (svr === null) {
    newServerCardV2(serverData);
    return
  }
  var props = ["address", "flavor", "motd", "name", "online", "players", "release", "running"];
  for (var i = 0; i < props.length; i++) {
    var ele = document.getElementById(props[i] + "_" + serverData.uuid);

    var val = "";
    if (props[i] === "online") {
      val = countNonEmpty(serverData.players);
    } else if (props[i] == "address") {
      val = hostname + ":" + serverData.port
    } else if (props[i] == "running") {
      val = runningToString(serverData.running);
    } else if (props[i] == "players") {
      val = listToVertical(serverData.players);
    } else {
      val = serverData[props[i]];
    }

    if (ele.innerHTML != val) {
      ele.innerHTML = val;
    }
  }
  updateCardActionButtons(serverData);
}

function updateCardActionButtons(serverData) {
  if (serverData.running === true) {
    document.getElementById("addOpIndicator_" + serverData.uuid).classList.remove("disabled");
    document.getElementById("weatherIndicator_" + serverData.uuid).classList.remove("disabled");
    document.getElementById("daytimeIndicator_" + serverData.uuid).classList.remove("disabled");
    document.getElementById("saveIndicator_" + serverData.uuid).classList.remove("disabled");
    document.getElementById("startIndicator_" + serverData.uuid).setAttribute("disabled", true);
    document.getElementById("stopIndicator_" + serverData.uuid).removeAttribute("disabled");

    if (serverData.whitelistenabled === true) {
      document.getElementById("whitelistPlayerIndicator_" + serverData.uuid).classList.remove("disabled");
    }
  } else {
    document.getElementById("whitelistPlayerIndicator_" + serverData.uuid).classList.add("disabled");
    document.getElementById("addOpIndicator_" + serverData.uuid).classList.add("disabled");
    document.getElementById("weatherIndicator_" + serverData.uuid).classList.add("disabled");
    document.getElementById("daytimeIndicator_" + serverData.uuid).classList.add("disabled");
    document.getElementById("saveIndicator_" + serverData.uuid).classList.add("disabled");
    document.getElementById("stopIndicator_" + serverData.uuid).setAttribute("disabled", true);
    document.getElementById("startIndicator_" + serverData.uuid).removeAttribute("disabled");
  }
  if (serverData.amowner === true) {
    document.getElementById("regenIndicator_" + serverData.uuid).removeAttribute("disabled");
    document.getElementById("deleteIndicator_" + serverData.uuid).classList.remove("disabled");
  } else {
    document.getElementById("regenIndicator_" + serverData.uuid).setAttribute("disabled", true);
    document.getElementById("deleteIndicator_" + serverData.uuid).classList.add("disabled");
  }
}

function countNonEmpty(arry) {
  if (arry === null) {
    return 0
  }
  var count = 0
  for (var i = 0; i < arry.length; i++) {
    if (arry[i] != "") {
      count++
    }
  }

  return count
}

function runningToString(val) {
  if (val === true) {
    return "Running"
  } else if (val === false) {
    return "Stopped"
  }
  return "Status Unknown"
}

function listToVertical(list) {
  var view = "";
  if (list === null) {
    return view
  }
  for (var i = 0; i < list.length; i++) {
    view += list[i] + "<br />";
  }
  return view
}

function sleep(milliseconds) {
  const date = Date.now();
  let currentDate = null;
  do {
    currentDate = Date.now();
  } while (currentDate - date < milliseconds);
}