const nav=document.querySelector('#side-navigation');
const topOfNav = nav.offsetTop;

function fixNav(){
	if(window.scrollY >= topOfNav){
		document.querySelector('.side-nav').classList.add('fixed-nav');
		document.querySelector('.documentation-content').classList.add('fixed-nav-documentation');
	} else{
		document.querySelector('.side-nav').classList.remove('fixed-nav');
		document.querySelector('.documentation-content').classList.remove('fixed-nav-documentation');

	}
}
window.addEventListener('scroll', fixNav);
