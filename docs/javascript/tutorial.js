var nodeList = $('.progressbar-bubble');

// Check that forward btn exists on page
var forwardBtn = $('#continue')[0];
forwardBtn ? completeState() : fullState()

/* Find index of element with active class to fill nodes
up until index of active class */
function completeState() {
  var current = nodeList.filter('.active'),
    index = nodeList.index(current);

  for (let i = 0; i < index; i++) {
    if (index == 0) return
    nodeList[i].className += 'completed';
  }
}

// Fill progress bar when at final, full state
function fullState() {
  for (let j = 0; j < nodeList.length; j++) {
    nodeList[j].className += 'completed';
  }
}
