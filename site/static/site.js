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
  var container = document.createElement("div");
  container.classList.add("col-sm-6", "col-lg-4", "mb-4", "newsitem");

  var card = document.createElement("div");
  card.classList.add("card", "shadow");

  itemDate = new Date(item.posted);
  if (isToday(itemDate)) {
    card.classList.add("today");
  }

  var cardimglink = document.createElement("a");
  cardimglink.setAttribute("href", item.preview.Link);
  cardimglink.setAttribute("target", "_blank");

  var cardimg = document.createElement("img");
  cardimg.classList.add("card-img-top");
  cardimg.setAttribute("src", item.preview.Images[0]);
  cardimg.setAttribute("class", "card-img-top");

  var cardbody = document.createElement("div");
  cardbody.classList.add("card-body", "bg-light");

  var cardtitle = document.createElement("h5");
  cardtitle.classList.add("card-title");
  cardtitle.innerText = item.preview.Title;

  var carddescr = document.createElement("p");
  carddescr.classList.add("card-text");
  carddescr.innerText = item.preview.Description;

  var cardfooter = document.createElement("p");
  cardfooter.classList.add("card-text");

  var footertext = document.createElement("small");
  footertext.classList.add("text-muted");
  footertext.innerText = item.posted+" ("+item.preview.Name+")";

  cardimglink.appendChild(cardimg);
  cardfooter.appendChild(footertext);

  cardbody.appendChild(cardtitle);
  cardbody.appendChild(carddescr);
  cardbody.appendChild(cardfooter);

  card.appendChild(cardimglink);
  card.appendChild(cardbody);

  container.appendChild(card);
  document.getElementById("news").appendChild(container);
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
  // Card Indicators
    var cardindicators = document.createElement("div");
    cardindicators.classList.add("mb-0");
    cardindicators.style.float = "right";

    var startindicator = document.createElement("a");
    startindicator.id = "startIndicator_"+item.uuid;
    startindicator.href = "#";
    startindicator.onclick = "startServer('"+item.uuid+"')";
    if (item.running === true) {
      startindicator.classList.add("hidden");
    }
    startindicator.innerHTML = '<i class="bi-play-fill text-success"></i>';
    cardindicators.appendChild(startindicator);

    var stopindicator = document.createElement("a");
    stopindicator.id = "stopIndicator_"+item.uuid;
    stopindicator.href = "#";
    stopindicator.onclick = "startServer('"+item.uuid+"')";
    if (item.running === false) {
      stopindicator.classList.add("hidden");
    }
    stopindicator.innerHTML = '<i class="bi-stop-fill text-warning"></i>';
    cardindicators.appendChild(stopindicator);

    if (item.amowner === true) {
      var deleteindicator = document.createElement("a");
      deleteindicator.id = "stopIndicator_"+item.uuid;
      deleteindicator.href = "#";
      deleteindicator.setAttribute("onclick", "deleteServer('"+item.name+"', '"+item.uuid+"')");
      if (item.running === true) {
        deleteindicator.classList.add("hidden");
      }
      deleteindicator.innerHTML = '<i class="bi-trash2 text-danger"></i>';
      cardindicators.appendChild(deleteindicator);
    }

  // Card Header
    var cardheader = document.createElement("h4");
    cardheader.classList.add("card-header");
    cardheader.appendChild(document.createTextNode(item.name));
    cardheader.appendChild(cardindicators);

  // Card Body
    var cardbody = document.createElement("div");
    cardbody.classList.add("card-body", "bg-light", "servercard");
    cardbody.appendChild(cardheader);

    var cardtitle = document.createElement("h5");
    cardtitle.classList.add("card-title");
    cardtitle.innerText = item.motd;
    cardbody.appendChild(cardtitle);
    cardbody.appendChild(document.createElement("br"));

    var cardtext = document.createElement("p");
    cardtext.classList.add("card-text");
    cardtext.innerHTML = `
    <strong>Flavor:</strong> `+item.flavor+`<br />
    <strong>Release:</strong> `+item.release+`<br />
    <strong>Port:</strong> `+item.port+`<br />
    <strong>Autostart:</strong> `+item.autostart+`<br />
    <strong>Ops:</strong> `+item.ops+`<br />
    <strong>Players:</strong> `+item.players+`<br />
    `;
    cardbody.appendChild(cardtext);

  // Card
    var card = document.createElement("div");
    card.classList.add("card", "shadow");
    card.appendChild(cardheader);
    card.appendChild(cardbody);

  // Container
    var container = document.createElement("div");
    container.id = item.uuid;
    container.classList.add("col-sm-6", "col-lg-4", "mb-4");
    container.appendChild(card);
  
  // Write container to page
  document.getElementById("servers").appendChild(container);
}
