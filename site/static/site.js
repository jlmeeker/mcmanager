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
        location.reload();
      } else {
        document.getElementById('dangerToastBody').innerText = "Error: "+this.responseText;
        toastList[1].show(); // dangerToast
      }
    }
  };
  xhttp.open("POST", "/v1/delete/"+id, true);
  xhttp.send();
}

function fetchServerStatuses() {
  var xhttp = new XMLHttpRequest();
  xhttp.onreadystatechange = function() {
    if (this.readyState == 4) {
      if (this.status == 200) {
        refreshServerCards(JSON.parse(this.responseText));
      } else if (this.status == 403) {
          // just keep silent, user isn't logged in
      } else {
          document.getElementById('dangerToastBody').innerText = "Error refreshing server statuses";
          toastList[1].show(); // dangerToast
      }
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
    document.getElementById('dangerToastBody').innerText = "Could not refesh card for "+status.name;
    toastList[1].show(); // dangerToast
    return;
  }

  document.getElementById("port_"+status.uuid).innerText = status.port;
  document.getElementById("autostart_"+status.uuid).innerText = status.autostart;
  document.getElementById("players_"+status.uuid).innerText = status.players;
  document.getElementById("flavor_"+status.uuid).innerText = status.flavor;
  document.getElementById("ops_"+status.uuid).innerText = status.ops;
  document.getElementById("release_"+status.uuid).innerText = status.release;

  if (status.running) {
    document.getElementById("startIndicator_"+status.uuid).classList.add("hidden");
    document.getElementById("stopIndicator_"+status.uuid).classList.remove("hidden");
  } else {
    document.getElementById("startIndicator_"+status.uuid).classList.remove("hidden");
    document.getElementById("stopIndicator_"+status.uuid).classList.add("hidden");
  }
}

function submitForm(loc, form){
  var data = new FormData(form);
  var xhttp = new XMLHttpRequest();
  xhttp.onreadystatechange = function() {
    if (this.readyState == 4) {
      if (this.status == 200) {
        location.reload();
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

function logout() {
  var xhttp = new XMLHttpRequest();
  xhttp.onreadystatechange = function() {
    if (this.readyState == 4) {
      if (this.status == 200) {
        location.reload();
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
      newNewsItem(item[1]);
    }
  }
}

function newNewsItem(item) {
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
  