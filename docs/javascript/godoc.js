// https://github.com/golang/tools/blob/master/godoc/static/godocs.js
'use strict';

function bindToggle(el) {
  $('.toggleButton', el).click(function() {
    if ($(this).closest(".toggle, .toggleVisible")[0] != el) {
      // Only trigger the closest toggle header.
      return;
    }

    if ($(el).is('.toggle')) {
      $(el).addClass('toggleVisible').removeClass('toggle');
    } else {
      $(el).addClass('toggle').removeClass('toggleVisible');
    }
  });
}

function bindToggles(selector) {
  $(selector).each(function(i, el) {
    bindToggle(el);
  });
}

function bindToggleLink(el, prefix) {
  $(el).click(function() {
    var href = $(el).attr('href');
    var i = href.indexOf('#'+prefix);
    if (i < 0) {
      return;
    }
    var id = '#' + prefix + href.slice(i+1+prefix.length);
    if ($(id).is('.toggle')) {
      $(id).find('.toggleButton').first().click();
    }
  });
}

function bindToggleLinks(selector, prefix) {
  $(selector).each(function(i, el) {
    bindToggleLink(el, prefix);
  });
}

// Custom fixxes to godoc-generated links within the document itself
function fixSourceLinks() {
    const REPOSITORY = $('meta[name=godoc_repository]').attr("content");
    const PACKAGE = $('meta[name=godoc_package]').attr("content");

    $('#content.documentation-content').find('a').each(function(){
        const href = $(this).attr('href');

        if (!href.startsWith('/src/')) {
            return;
        }

        // Fix links of 'type' headers
        let newHref = href.replace('/src/target/', `https://${REPOSITORY}/tree/master/${PACKAGE}/`);

        // Fix links to package file listing
        newHref = newHref.replace(`/src/${REPOSITORY}/`, `https://${REPOSITORY}/tree/master/`);

        // Strip query params
        newHref = newHref.replace(/\?.*/, '');

        // Uncomment the following line for debugging
        // console.log(newHref);

        // Set the new link
        $(this).attr('href', newHref);
    });
}

$(document).ready(function() {
  bindToggles(".toggle");
  bindToggles(".toggleVisible");
  bindToggleLinks(".exampleLink", "example_");
  bindToggleLinks(".overviewLink", "");
  bindToggleLinks(".examplesLink", "");
  bindToggleLinks(".indexLink", "");

  fixSourceLinks();
});
