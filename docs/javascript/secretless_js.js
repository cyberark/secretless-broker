const nav=document.querySelector('#side-navigation');
const topOfNav = nav.offsetTop;
const docConClass = document.querySelector('.documentation-content').classList;

function fixNav(){
	if(window.scrollY >= topOfNav){
		document.querySelector('.side-nav').classList.add('fixed-nav');
		docConClass.add('fixed-nav-documentation');
	} else{
		document.querySelector('.side-nav').classList.remove('fixed-nav');
		docConClass.remove('fixed-nav-documentation');

	}
}
window.addEventListener('scroll', fixNav);
