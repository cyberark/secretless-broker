// toggleFixedNav
// This section fixes the documentation sidebar in place when the page is
// scrolled, by toggling on/off the fixed-nav class
document.addEventListener("DOMContentLoaded", function() {
		nav = document.querySelector('#side-navigation');
		topOfNav = nav.offsetTop;
		navHeight = nav.offsetHeight;

		footer = document.querySelector('footer')
		footerHeight = footer.offsetTop;

		console.log('footer: ' + footerHeight)

		docConClass = document.querySelector('.documentation-content').classList;

		window.addEventListener('scroll', function() { toggleFixedNav(); });
	});

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

// showTab
// Traverses up the DOM and finds `tab-content` class elements which share a
// common root with the `event` source. Once found, all `tab-content` elements
// are hidden, except for the element whose `id` matches `tabName`.
function showTab(event, tabName) {
	// Toggle `active` class on sibling buttons
	var srcElement = event.srcElement;
	const parentElement = srcElement.parentElement;
	const srcSiblings = parentElement.getElementsByClassName('button');
	for (var i = 0; i < srcSiblings.length; ++i) {
		srcSiblings[i].classList.remove('active');
	}
	srcElement.classList.add('active');

	// Find a common root between the event source and `tab-content` elements
	var closestTabs = [];
	while (srcElement != null) {
		const tabs = srcElement.getElementsByClassName('tab-content');
		if (tabs.length > 0) {
			closestTabs = tabs;
			break;
		}

		srcElement = srcElement.parentElement;
	}

	// Toggle `tab-content` visibility
	for (var i = 0; i < closestTabs.length; ++i) {
		const tab = closestTabs[i];
		const display = tab.id === tabName ? 'block' : 'none';

		tab.style.display = display;
	}
}
