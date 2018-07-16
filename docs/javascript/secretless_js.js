const nav=document.querySelector('#side-navigation');
const topOfNav = nav.offsetTop;
const docConClass = document.querySelector('.documentation-content').classList;

function fixNav(){
	if(window.scrollY >= topOfNav){
		document.querySelector('.side-nav').classList.add('fixed-nav');
		docConClass.add('fixed-nav-documentation');
	} else{
		document.querySelector('.side-nav').classList.remove('fixed-nav');

		footer = document.querySelector('footer')
		footerHeight = footer.offsetTop;

		console.log('footer: ' + footerHeight)

		docConClass = document.querySelector('.documentation-content').classList;

		window.addEventListener('scroll', function() { toggleFixedNav(); });
	});

function toggleFixedNav() {
	const sideNav = document.querySelector('.side-nav');
	const navBottom = window.scrollY + navHeight
	console.log(navBottom);
	if (navBottom >= footerHeight) {
		console.log("i am overflowing");
		sideNav.classList.remove('fixed-nav');
		docConClass.remove('fixed-nav-documentation');
	}
	if (window.scrollY >= topOfNav) {
		sideNav.classList.add('fixed-nav');
		docConClass.add('fixed-nav-documentation');
	}
	else {
		sideNav.classList.remove('fixed-nav');
		docConClass.remove('fixed-nav-documentation');

	}
}
window.addEventListener('scroll', fixNav);


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