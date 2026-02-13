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

// Confirm incomplete vote submission (two-click pattern)
(function () {
    document.querySelectorAll('form[data-confirm-incomplete]').forEach(function (form) {
        var btn = form.querySelector('button[type="submit"]');
        if (!btn) return;
        var originalText = btn.textContent.trim();
        var confirmText = form.dataset.confirmIncomplete;
        var armed = false;

        function hasEmpty() {
            var inputs = form.querySelectorAll('input[type="hidden"][name^="vote-"]');
            var empty = false;
            inputs.forEach(function (input) {
                if (!input.getAttribute('form') && input.value === '') {
                    empty = true;
                }
            });
            return empty;
        }

        function reset() {
            if (!armed) return;
            armed = false;
            btn.textContent = originalText;
            btn.classList.remove('bg-amber-500', 'hover:bg-amber-600');
            btn.classList.add('bg-primary-500', 'hover:bg-primary-600');
        }

        form.addEventListener('submit', function (e) {
            var nameInput = form.querySelector('input[name="name"]');
            if (!nameInput || !nameInput.value.trim()) return;
            if (!hasEmpty()) return;
            if (armed) { armed = false; return; }
            e.preventDefault();
            armed = true;
            btn.textContent = confirmText;
            btn.classList.remove('bg-primary-500', 'hover:bg-primary-600');
            btn.classList.add('bg-amber-500', 'hover:bg-amber-600');
        });

        // Reset when the user interacts with vote buttons
        form.querySelectorAll('.vote-btn').forEach(function (voteBtn) {
            voteBtn.addEventListener('click', function () { reset(); });
        });
    });
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

// Scroll fade + arrow indicators for horizontally-overflowing containers
(function () {
    var chevronSvg = '<svg class="size-4" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M8.22 5.22a.75.75 0 0 1 1.06 0l4.25 4.25a.75.75 0 0 1 0 1.06l-4.25 4.25a.75.75 0 0 1-1.06-1.06L11.94 10 8.22 6.28a.75.75 0 0 1 0-1.06Z" clip-rule="evenodd"/></svg>';
    var arrowClasses = 'flex size-7 items-center justify-center rounded-full bg-white text-text-500 shadow-sm ring-1 ring-background-200 transition-opacity duration-300 hover:bg-background-50 hover:text-text-700 dark:bg-background-100 dark:ring-background-300 dark:hover:bg-background-200';

    document.querySelectorAll('[data-scroll-fade]').forEach(function (el) {
        var wrapper = document.createElement('div');
        wrapper.className = 'relative';
        el.parentNode.insertBefore(wrapper, el);
        wrapper.appendChild(el);

        // Prevent iOS bounce/overscroll on the scroll container
        el.style.overscrollBehaviorX = 'contain';

        // Fade gradients
        var fadeRight = document.createElement('div');
        fadeRight.className = 'pointer-events-none absolute inset-y-0 right-0 w-12 rounded-r-lg bg-linear-to-l from-white dark:from-background-100 transition-opacity duration-300';
        wrapper.appendChild(fadeRight);

        var fadeLeft = document.createElement('div');
        fadeLeft.className = 'pointer-events-none absolute inset-y-0 left-0 w-12 rounded-l-lg bg-linear-to-r from-white dark:from-background-100 transition-opacity duration-300';
        wrapper.appendChild(fadeLeft);

        // Arrow button group pinned to top-right of the table
        var arrowGroup = document.createElement('div');
        arrowGroup.className = 'absolute -top-9 right-0 flex items-center gap-1 transition-opacity duration-300';
        wrapper.appendChild(arrowGroup);

        var arrowLeft = document.createElement('button');
        arrowLeft.type = 'button';
        arrowLeft.setAttribute('aria-label', 'Scroll left');
        arrowLeft.className = arrowClasses + ' rotate-180';
        arrowLeft.innerHTML = chevronSvg;
        arrowGroup.appendChild(arrowLeft);

        var arrowRight = document.createElement('button');
        arrowRight.type = 'button';
        arrowRight.setAttribute('aria-label', 'Scroll right');
        arrowRight.className = arrowClasses;
        arrowRight.innerHTML = chevronSvg;
        arrowGroup.appendChild(arrowRight);

        arrowRight.addEventListener('click', function () {
            var max = el.scrollWidth - el.clientWidth;
            var target = Math.min(el.scrollLeft + el.clientWidth * 0.75, max);
            el.scrollTo({ left: target, behavior: 'smooth' });
        });

        arrowLeft.addEventListener('click', function () {
            var target = Math.max(el.scrollLeft - el.clientWidth * 0.75, 0);
            el.scrollTo({ left: target, behavior: 'smooth' });
        });

        function update() {
            var overflows = el.scrollWidth > el.clientWidth;
            var atStart = el.scrollLeft <= 1;
            var atEnd = el.scrollLeft + el.clientWidth >= el.scrollWidth - 1;

            fadeRight.classList.toggle('opacity-0', !overflows || atEnd);
            fadeLeft.classList.toggle('opacity-0', !overflows || atStart);

            // Show/hide the entire arrow group
            arrowGroup.classList.toggle('opacity-0', !overflows);
            arrowGroup.classList.toggle('pointer-events-none', !overflows);

            // Dim individual arrows at scroll bounds
            arrowRight.classList.toggle('opacity-30', overflows && atEnd);
            arrowRight.classList.toggle('pointer-events-none', atEnd);
            arrowLeft.classList.toggle('opacity-30', overflows && atStart);
            arrowLeft.classList.toggle('pointer-events-none', atStart);
        }

        el.addEventListener('scroll', update);
        window.addEventListener('resize', update);
        update();
    });
})();
