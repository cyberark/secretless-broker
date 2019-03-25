var items = document.querySelectorAll(".item")
remainingItems = [].slice.call(items)

var activeItems = []
var fBtn = document.getElementById("fBtn")
var bBtn = document.getElementById("bBtn")

// update stack if user advances
function stepForward() {
  const nextItem = remainingItems.pop()
  if (!nextItem) return
  activeItems.push(nextItem)
  // render update
  renderItems()
}

// update stack if user regresses
function stepBack() {
  const previousItem = activeItems.pop()
  if (!previousItem) return
  remainingItems.push(previousItem)
  if (activeItems.length < 1) return
  renderItems()
}

function renderItems() {
  let zIndex = 1
  activeItems.forEach(item => {
    item.style.visibility = 'visible'
    item.style.zIndex = zIndex
    zIndex += 1
  })

  remainingItems.forEach(item => {
    item.style.visibility = 'hidden'
  })
}

fBtn.addEventListener("click", stepForward)
bBtn.addEventListener("click", stepBack)

stepForward()
