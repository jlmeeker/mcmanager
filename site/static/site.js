// Server Actions
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
      if (this.status == 200) {
        document.getElementById('successToastBody').innerText = "Successfully "+action+"ed the server";
        toastList[0].show(); // successToast
      } else {
        document.getElementById('dangerToastBody').innerText = "Error: "+this.responseText;
        toastList[1].show(); // dangerToast
      }
      fetchServers();
    }
  };
  xhttp.open("POST", "/v1/"+action+"/"+id, true);
  xhttp.send();
}

function deleteServer(name, id) {
  var r = confirm("Delete "+name+"?");
  if ( r === false) {
    return false;
  }

  var xhttp = new XMLHttpRequest();
  xhttp.onreadystatechange = function() {
    if (this.readyState == 4) {
      if (this.status == 200) {
        document.getElementById('successToastBody').innerText = "Server successfully deleted";
        toastList[0].show(); // successToast
        fetchServers();
      } else {
        document.getElementById('dangerToastBody').innerText = "Error: "+this.responseText;
        toastList[1].show(); // dangerToast
      }
    }
  };
  xhttp.open("POST", "/v1/delete/"+id, true);
  xhttp.send();
}

// Modal Actions
function closeModal(id) {
  var myModalEl = document.getElementById(id);
  var modal = bootstrap.Modal.getInstance(myModalEl)
  modal.hide();
}

// All Forms (new server, login etc.)
function submitForm(loc, form){
  var data = new FormData(form);
  var xhttp = new XMLHttpRequest();
  xhttp.onreadystatechange = function() {
    if (this.readyState == 4) {
      if (this.status == 200) {
        document.getElementById('successToastBody').innerText = "Success";
        toastList[0].show(); // successToast
        if (loc == "/v1/create") {
          closeModal('newServerModal');
          fetchServers();
        } else if (loc == "/v1/login") {
          document.getElementById('newServerIcon').classList.remove("hidden");
          document.getElementById('logOutButton').classList.remove("hidden");
          document.getElementById('logInButton').classList.add("hidden");
          var replyObj = JSON.parse(this.responseText);
          document.getElementById('playerName').innerText = replyObj.playername;
          closeModal('logInModal');

          if (replyObj.page == "servers") {
            fetchServers();
          }
        }
        form.reset();
      } else {
        document.getElementById('dangerToastBody').innerText = "Error: "+this.responseText;
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
  card.id = item.uuid;
  card.innerHTML =  `
  <div class="card shadow">
    <a href="`+item.preview.Link+`" target="_blank">
      <img class="card-img-top" src="`+item.preview.Images[0]+`"></a>
      <div class="card-body bg-light">
        <h5 class="card-title">`+item.preview.Title+`</h5>
        <p class="card-text">`+item.preview.Description+`</p>
        <p class="card-text">
          <small class="text-muted">`+item.posted+` (`+item.preview.Name+`)</small>
        </p>
      </div>
    </div>
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
  card.classList.add("col-sm-6", "col-lg-4", "mb-4");
  card.id = item.uuid;
  card.innerHTML = `
    <div class="card shadow">
        <h4 class="card-header">`+item.name+`
            <div class="mb-0" style="float: right;">
              <a id="startIndicator_`+item.uuid+`" href="#" class="hidden" onClick="startServer('`+item.uuid+`')">
                <i class="bi-play-fill text-success"></i>
              </a>
              <a id="stopIndicator_`+item.uuid+`" href="#" class="hidden" onClick="stopServer('`+item.uuid+`')">
                <i class="bi-stop-fill text-warning"></i>
              </a>
              <a id="deleteIndicator_`+item.uuid+`" href="#" onclick="deleteServer('`+item.name+`', '`+item.uuid+`')" class="hidden">
                <i class="bi-trash2 text-danger"></i>
              </a>
            </div>
        </h4>
        <div class="card-body bg-light servercard">
            <h5 class="card-title">`+item.motd+`</h5><br>
            <p class="card-text">
              <strong>Flavor:</strong> `+item.flavor+`<br>
              <strong>Release:</strong> `+item.release+`<br>
              <strong>Port:</strong> `+item.port+`<br>
              <strong>Autostart:</strong> `+item.autostart+`<br>
              <strong>Ops:</strong> `+item.ops+`<br>
              <strong>Players:</strong> `+item.players+`<br>
            </p>
        </div>
    </div>
  `;
  document.getElementById("servers").appendChild(card);
  if (item.running === true) {
    document.getElementById("stopIndicator_"+item.uuid).classList.remove("hidden");
  } else {
    document.getElementById("startIndicator_"+item.uuid).classList.remove("hidden");
  }
  if (item.amowner === true) {
    document.getElementById("deleteIndicator_"+item.uuid).classList.remove("hidden");
  }
}
