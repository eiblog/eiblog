{{define "ana_js"}}
! function(e, n, o, t) {var a = e.screen,r = encodeURIComponent,i = ["tid=UA-77251712-1", "dl=" + r(n.URL), "dt=" + r(n.title), "dr=" + r(n.referrer), "dp=" + r(t.pathname), "ul=" + (o.language || o.browserLanguage).toLowerCase(), "sd=" + a.colorDepth + "-bit", "sr=" + a.width + "x" + a.height, "_=" + +new Date],c = "?" + i.join("&");e.__beacon_img = new Image, e.__beacon_img.src = "/beacon.html" + c }(window, document, navigator, location);
{{end}}
