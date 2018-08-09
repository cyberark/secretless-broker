/* Loop through all dropdown buttons and allow for toggling dropdown content when clicked
This allows freedom to have users select multiple dropdown carets and expose dropdown menu content*/
var dropdownOption = document.querySelectorAll(".dropdown-btn");

dropdownOption.forEach(er => er.addEventListener("click", function () {
	this.classList.toggle("active");
	var dropdownContent = this.nextElementSibling;
	if (dropdownContent.style.display === "block") {
		dropdownContent.style.display = "none";
		console.log(this.nextElementSibling);
	} else {
		dropdownContent.style.display = "block";
	}
}));

// Rotate arrow in side navbar if 3rd tier is opened or closed
$(".dropdown-btn").click(function(){
	 $(this).find(".fa-angle-right").toggleClass("rotatingArrow");
})