// toggleFixedNav
// This section fixes the documentation sidebar in place when the page is
// scrolled, by toggling on/off the fixed-nav class
document.addEventListener("DOMContentLoaded", function() {
		nav = document.querySelector('#side-navigation');

		if (nav !== null) {
			topOfNav = nav.offsetTop;
			navHeight = nav.offsetHeight;

			docConClass = document.querySelector('.documentation-content').classList;

			window.addEventListener('scroll', function() { toggleFixedNav(); });
		}
	}
);

function toggleFixedNav() {
	const sideNav = document.querySelector('.side-nav');
	const navBottom = window.scrollY + navHeight

	if (window.scrollY >= topOfNav) {
		sideNav.classList.add('fixed-nav');
		docConClass.add('fixed-nav-documentation');
	} else {
		sideNav.classList.remove('fixed-nav');
		docConClass.remove('fixed-nav-documentation');
	}
}
