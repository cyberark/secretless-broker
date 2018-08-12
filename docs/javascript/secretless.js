/* Loop through all dropdown buttons and allow for toggling dropdown content when clicked
This allows freedom to have users select multiple dropdown carets and expose dropdown menu content*/
var dropdownOption = document.querySelectorAll(".dropdown-btn");

dropdownOption.forEach(er => er.addEventListener("click", function () {
	this.classList.toggle("active");
	var dropdownContent = this.nextElementSibling;
	if (dropdownContent.style.display === "block") {
		dropdownContent.style.display = "none";
	} else {
		dropdownContent.style.display = "block";
	}
}));

// Rotate arrow in side navbar if 3rd tier is opened or closed
$(".dropdown-btn").click(function(){
	 $(this).find(".fa-angle-right").toggleClass("rotatingArrow");
})

// 	Fires upon page reload
$(document).ready ( function(){
	// Retrieves the category for each dropdown
   console.log(extractParmsURL()[5].split("%")[0]);

})

// Retrieves URL and extracts the correct category
function extractParmsURL(e) {
	var urlParams = new URLSearchParams(window.location.href);
	console.log(urlParams.toString());
	var key = urlParams.toString().split("2F");

	checkState(e);
	return key;
	// urlParams.append("changes", 4);
	// window.history.replaceState({}, urlParams.append("changes", 4));
	// window.history.pushState(("changes", 4),1);
	// console.log(window.history);
	// console.log(urlParams.toString());

	// var key = urlParams.toString();
	// console.log(key);
	// var decomposedKey = key.split("F");
	// console.log(decomposedKey);
	// var test=key.replace("%", "s");
	
	// console.log(test);
}

function checkState(e){
	// console.log(this.nextElementSibling.classList.toggle("test"));
	// this.classList.toggle("test");
	console.log(e);
	// console.log(window.event.path.classList.toggle("test")); //put on subnav-subitem
	// console.log(window.event.path[1].classList.toggle("test"));
	// console.log(window.event.path[1].classList);
	// console.log(event.target.querySelector(''))
}
