{{define "ana_js"}}
(function(i, s, o, g, r, a, m) {
    i['GoogleAnalyticsObject'] = r;
    i[r] = i[r] || function() {
        (i[r].q = i[r].q || []).push(arguments)
    }, i[r].l = 1 * new Date();
    a = s.createElement(o),
        m = s.getElementsByTagName(o)[0];
    a.async = 1;
    a.src = g;
    m.parentNode.insertBefore(a, m)
})(window, document, 'script', 'https://o7msui8ho.qnssl.com/js/analytics.js', 'ga');

ga('create', 'UA-77251712-1', 'auto');
ga('send', 'pageview');
! function(e, n, o, t) {
    var a = e.screen,
        r = encodeURIComponent,
        i = ["tid=UA-5422922-2", "dl=" + r(n.URL), "dt=" + r(n.title), "dr=" + r(n.referrer), "dp=" + r(t.pathname), "ul=" + (o.language || o.browserLanguage).toLowerCase(), "sd=" + a.colorDepth + "-bit", "sr=" + a.width + "x" + a.height, "_=" + +new Date],
        c = "?" + i.join("&");
    e.__beacon_img = new Image, e.__beacon_img.src = "/beacon.html" + c
}(window, document, navigator, location);
{{end}}
