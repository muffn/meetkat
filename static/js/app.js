// Settings popover (theme + language switcher)
(function () {
    var settingsBtn = document.getElementById('settings-btn');
    var popover = document.getElementById('settings-popover');
    var themeButtons = document.querySelectorAll('.theme-btn');
    var langButtons = document.querySelectorAll('.lang-btn');
    var checkedClasses = ['bg-white', 'shadow-sm', 'ring-1', 'ring-text-900/10', 'dark:bg-background-300', 'dark:ring-transparent'];
    var open = false;

    function togglePopover() {
        open = !open;
        if (open) {
            popover.classList.remove('opacity-0', 'scale-95', 'pointer-events-none');
            popover.classList.add('opacity-100', 'scale-100');
        } else {
            popover.classList.remove('opacity-100', 'scale-100');
            popover.classList.add('opacity-0', 'scale-95', 'pointer-events-none');
        }
    }

    function closePopover() {
        if (!open) return;
        open = false;
        popover.classList.remove('opacity-100', 'scale-100');
        popover.classList.add('opacity-0', 'scale-95', 'pointer-events-none');
    }

    function getTheme() {
        var theme = localStorage.theme;
        if (!theme) {
            var m = document.cookie.match(/(?:^|;\s*)meetkat_theme=(\w+)/);
            if (m) theme = m[1];
        }
        return theme || '';
    }

    function applyTheme() {
        var theme = getTheme();
        document.documentElement.classList.toggle(
            'dark',
            theme === 'dark' || (!theme && window.matchMedia('(prefers-color-scheme: dark)').matches)
        );
    }

    function updateThemeButtons() {
        var current = getTheme() || 'system';
        themeButtons.forEach(function (btn) {
            var isActive = btn.dataset.theme === current;
            btn.setAttribute('aria-checked', String(isActive));
            if (isActive) {
                btn.classList.add.apply(btn.classList, checkedClasses);
            } else {
                btn.classList.remove.apply(btn.classList, checkedClasses);
            }
        });
    }

    function getCurrentLang() {
        return document.documentElement.lang || 'en';
    }

    function updateLangButtons() {
        var current = getCurrentLang();
        langButtons.forEach(function (btn) {
            var isActive = btn.dataset.lang === current;
            btn.setAttribute('aria-checked', String(isActive));
            if (isActive) {
                btn.classList.add.apply(btn.classList, checkedClasses);
            } else {
                btn.classList.remove.apply(btn.classList, checkedClasses);
            }
        });
    }

    settingsBtn.addEventListener('click', function (e) {
        e.stopPropagation();
        togglePopover();
    });

    themeButtons.forEach(function (btn) {
        btn.addEventListener('click', function () {
            if (btn.dataset.theme === 'system') {
                localStorage.removeItem('theme');
                document.cookie = 'meetkat_theme=; path=/; max-age=0';
            } else {
                localStorage.theme = btn.dataset.theme;
                document.cookie = 'meetkat_theme=' + btn.dataset.theme + '; path=/; max-age=31536000; SameSite=Lax';
            }
            applyTheme();
            updateThemeButtons();
            closePopover();
        });
    });

    langButtons.forEach(function (btn) {
        btn.addEventListener('click', function () {
            var lang = btn.dataset.lang;
            if (lang === getCurrentLang()) {
                closePopover();
                return;
            }
            var params = new URLSearchParams(window.location.search);
            params.set('lang', lang);
            window.location.href = window.location.pathname + '?' + params.toString();
        });
    });

    document.addEventListener('click', function (e) {
        if (open && !popover.contains(e.target) && !settingsBtn.contains(e.target)) {
            closePopover();
        }
    });

    updateThemeButtons();
    updateLangButtons();
})();

// Vote button toggles (yes/no)
(function () {
    document.querySelectorAll('.vote-btn').forEach(function (btn) {
        btn.addEventListener('click', function () {
            var td = btn.closest('td');
            var hidden = td.querySelector('input[type="hidden"]');
            var buttons = td.querySelectorAll('.vote-btn');

            hidden.value = btn.dataset.value;

            buttons.forEach(function (b) {
                b.classList.remove('bg-green-100', 'text-green-600', 'bg-red-100', 'text-red-500');
                b.classList.add('bg-background-50', 'text-background-300');
                b.setAttribute('aria-pressed', 'false');
            });

            btn.classList.remove('bg-background-50', 'text-background-300');
            if (btn.dataset.value === 'yes') {
                btn.classList.add('bg-green-100', 'text-green-600');
            } else {
                btn.classList.add('bg-red-100', 'text-red-500');
            }
            btn.setAttribute('aria-pressed', 'true');
        });
    });
})();

// Scroll fade indicator for horizontally-overflowing containers
(function () {
    document.querySelectorAll('[data-scroll-fade]').forEach(function (el) {
        var wrapper = document.createElement('div');
        wrapper.className = 'relative';
        el.parentNode.insertBefore(wrapper, el);
        wrapper.appendChild(el);

        var fade = document.createElement('div');
        fade.className = 'pointer-events-none absolute inset-y-0 right-0 w-12 rounded-r-lg bg-linear-to-l from-white dark:from-background-100 transition-opacity duration-300';
        wrapper.appendChild(fade);

        function update() {
            var overflows = el.scrollWidth > el.clientWidth;
            var atEnd = el.scrollLeft + el.clientWidth >= el.scrollWidth - 1;
            fade.classList.toggle('opacity-0', !overflows || atEnd);
        }

        el.addEventListener('scroll', update);
        window.addEventListener('resize', update);
        update();
    });
})();
