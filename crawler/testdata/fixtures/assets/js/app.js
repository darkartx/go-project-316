// Main application JavaScript
'use strict';

(function() {
    // Initialize app
    document.addEventListener('DOMContentLoaded', function() {
        initNavigation();
        initLazyLoading();
        trackPageView();
    });

    function initNavigation() {
        var navLinks = document.querySelectorAll('nav a');
        navLinks.forEach(function(link) {
            link.addEventListener('click', function(e) {
                console.log('Navigation click:', this.href);
            });
        });
    }

    function initLazyLoading() {
        var images = document.querySelectorAll('img[data-src]');
        images.forEach(function(img) {
            img.src = img.dataset.src;
        });
    }

    function trackPageView() {
        console.log('Page viewed:', window.location.pathname);
    }
})();