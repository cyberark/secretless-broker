/* Loop through all dropdown buttons and allow for toggling dropdown content when clicked
This allows freedom to have users select multiple dropdown carets and expose dropdown menu content*/
var dropdownOption = document.querySelectorAll(".dropdown-btn");

dropdownOption.forEach(er => er.addEventListener("click", function () {
	this.classList.toggle("active");
	var dropdownContent = this.nextElementSibling;
	if (!dropdownContent.classList.toggle("navbar-open")) {
		this.classList.add("navbar-open"); 
	} else{
		this.classList.remove("navbar-open");
	}
}));

// Rotate arrow in side navbar if 3rd tier is opened or closed
dropdownOption.forEach(er => er.addEventListener("click", function () {
	this.children[1].classList.toggle("rotatingArrow");
}));
