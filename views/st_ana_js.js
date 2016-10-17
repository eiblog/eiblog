{{define "ana_js"}}
!function(e,n,o){var t=e.screen,a=encodeURIComponent,r=["dt="+a(n.title),"dr="+a(n.referrer),"ul="+(o.language||o.browserLanguage),"sd="+t.colorDepth+"-bit","sr="+t.width+"x"+t.height,"_="+ +new Date],i="?"+r.join("&");e.__beacon_img=new Image,e.__beacon_img.src="/beacon.html"+i}(window,document,navigator,location);
{{end}}