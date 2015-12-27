angular.module('app', [])
  .filter('percentage', ['$filter', function ($filter) {
    return function (input, decimals) {
      return $filter('number')(input * 100, decimals) + '%';
    };
  }])
  .controller('mainController', function($scope, $location) {

    //$scope.events = [{Hostname: "smwa.me", Apps: [{Appname: "Apache2", Severities: [{Severity: 3, Count: 4}]}]}];
    $scope.events = [];
    $scope.errors = [];

    var severityMap = {
        "0": "Emergency",
        "1": "Alert",
        "2": "Critical",
        "3": "Error",
        "4": "Warning",
        "5": "Notice",
        "6": "Informational",
        "7": "Debug"
    };

    var AudioSets = {
        "StarTrek": {
            "Events": {
                "0": "StarTrek/tng_torpedo_clean.mp3",
                "1": "StarTrek/tos_ship_phaser_1.mp3",
                "2": "StarTrek/tos_tricoder_alert.mp3",
                "3": "StarTrek/tos_chirp_5.mp3",
                "4": "StarTrek/computerbeep_22.mp3",
                "5": "StarTrek/tng_viewscreen_off.mp3",
                "6": "StarTrek/hypospray2_clean.mp3",
                "7": "StarTrek/communications_start_transmission.mp3",
            }
        }
    };

    $scope.CurrentAudioSet = "StarTrek";

    function connect() {
        var protocol = "ws";
        if (window.location.protocol == "https:") {
            protocol = "wss";
        }
        $scope.websocketconnection = new WebSocket("ws://"+$location.host()+":"+$location.port()+"/websocket");
        $scope.websocketconnection.onopen = function() {
            $scope.errors = [];
            $scope.$apply();
        }
        $scope.websocketconnection.onclose = function() {
            $scope.errors = ["Connection failed, trying to reconnect"];
            setTimeout(connect, 999);
            $scope.$apply();
        }
        $scope.websocketconnection.onmessage = function(evt){ onMessage(evt); };
    }

    function onMessage(evt) {
        evt = JSON.parse(evt.data);
        if(evt.Appname == "") {
            evt.Appname = "N/A";
        }
        if (severityMap.hasOwnProperty(evt.Severity)) {
            evt.SeverityText = severityMap[evt.Severity];
        }
        //Host
        var hostid = -1;
        for (var i = 0; i < $scope.events.length; i++) {
            if ($scope.events[i].Hostname == evt.Hostname) {
                hostid = i;
            }
        }
        if (hostid < 0) {
            $scope.events.push({Hostname: evt.Hostname, CpuUsage: 0, CpuVolume: 1, MemoryUsage: 0, MemoryVolume: 1, Apps: []});
            hostid = $scope.events.length - 1;
        }

        //states
        if (evt.Appname == "cpu_state") {
            $scope.events[hostid].CpuUsage = parseFloat(evt.Intensity);
            //affect sound
            $scope.$apply();
            return;
        }
        if (evt.Appname == "memory_state") {
            $scope.events[hostid].MemoryUsage = parseFloat(evt.Intensity);
            //affect sound
            $scope.$apply();
            return;
        }

        //App
        var appid = -1;
        for (var i = 0; i < $scope.events[hostid].Apps.length; i++) {
            if ($scope.events[hostid].Apps[i].Appname == evt.Appname) {
                appid = i;
            }
        }
        if (appid < 0) {
            $scope.events[hostid].Apps.push({Appname: evt.Appname, Severities: []});
            appid = $scope.events[hostid].Apps.length - 1;
        }

        //Severities
        var app = $scope.events[hostid].Apps[appid]
        var sevid = -1;
        for (var i = 0; i < app.Severities.length; i++) {
            if (app.Severities[i].Severity == evt.SeverityText) {
                sevid = i;
            }
        }
        if (sevid < 0) {
            app.Severities.push({Severity: evt.SeverityText, Count: 0, Volume: 1.0});
            sevid = app.Severities.length - 1;
        }

        //increase counter
        app.Severities[sevid].Count++;
        playEventSound(evt, app.Severities[sevid].Volume);
        $scope.$apply();
    }

    $scope.errors.push("Not connected yet");
    connect();


    //$scope.audio = new window.AudioContext();
    var AudiosPoolPositionTracker = 0;
    var AudiosPoolSize = 5;
    var AudiosPool = [];
    for (var i = 0; i < AudiosPoolSize; i++) {
        AudiosPool.push(new Audio());
    }

    function playEventSound(evt, volume) {
        var sound = AudiosPool[AudiosPoolPositionTracker++];
        sound.src = "audio/"+AudioSets[$scope.CurrentAudioSet].Events[evt.Severity];
        sound.volume = volume;
        sound.play();
        AudiosPoolPositionTracker = AudiosPoolPositionTracker % 5;
        /*
        freqmap = {
            "0": 261.626,
            "1": 329.628,
            "2": 391.995,
            "3": 523.251,
            "4": 659.255,
            "5": 783.991,
            "6": 1046.502,
            "7": 1567.982
        };
        playNote(freqmap[evt.Severity], volume);
        */
    }

    function playNote(freq, volume) {
        var audio = $scope.audio;
        var attack = 10;
        var decay = 250
        var gain = audio.createGain();
        var osc = audio.createOscillator();

        gain.connect(audio.destination);
        gain.gain.setValueAtTime(0, audio.currentTime);
        gain.gain.linearRampToValueAtTime(volume, audio.currentTime + attack / 1000);
        gain.gain.linearRampToValueAtTime(0, audio.currentTime + decay / 1000);

        osc.frequency.value = freq;
        osc.type = "square";
        osc.connect(gain);
        osc.start(0);

        setTimeout(function() {
            osc.stop(0);
            osc.disconnect(gain);
            gain.disconnect(audio.destination);
        }, decay);
    }

});
