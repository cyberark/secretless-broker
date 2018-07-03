# CyberArk Secretless Website

This website is generated using Jeykll. We've documented our uses and best practices in a [separate file](jekyll-structure.md).

### Prerequisites
To get the site up and running locally on your computer, ensure you have:
	1. Ruby version 2.1.0 or higher (check by running `ruby -v`)

### Run Locally
To construct:
	1. Git clone the secretless repository https://github.com/conjurinc/secretless.git
	2. Switch to branch 87-jekyll-website (temporary)
	3. Navigate to the 'docs' directory of pulled repository
	4. Run the following command:
	```
	sh-session $ bundle exec jekyll serve
	```
	5. Preview Jekyll site locally in web browser by either running `open localhost:4000` or manually navigating to http://localhost:4000