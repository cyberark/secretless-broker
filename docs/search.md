---
title: Search
id: search
layout: subpages
description: Search Secretless
---

<!-- Algolia Search -->
<div class="search-wrap">
  <div id="search-searchbar"></div>
  <div class="plop" id="search-hits"></div>
  <div id="current-refined-values">
      <!-- CurrentRefinedValues widget will appear here -->
    </div>
    
    <div id="clear-all"><!-- ClearAll widget --></div>
    <div id="pagination"><!-- Pagination widget --></div>
  {% include algolia.html %}
</div>