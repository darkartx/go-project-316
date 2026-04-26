// Vendor JavaScript library
(function(global) {
    'use strict';

    // Mini utility library
    var lib = {
        version: '1.0.0',

        get: function(url, callback) {
            var xhr = new XMLHttpRequest();
            xhr.open('GET', url);
            xhr.onreadystatechange = function() {
                if (xhr.readyState === 4 && xhr.status === 200) {
                    callback(null, xhr.responseText);
                }
            };
            xhr.send();
        },

        ready: function(fn) {
            if (document.readyState !== 'loading') {
                fn();
            } else {
                document.addEventListener('DOMContentLoaded', fn);
            }
        }
    };

    global.lib = lib;
})(window);