# CyberArk Secretless Website

This website is generated using Jekyll. We've documented our uses and best practices in a [separate file](jekyll-structure.md).

### Prerequisites
To get the site up and running locally on your computer, ensure you have:
1. Ruby version 2.1.0 or higher (check by running `ruby -v`)
2. Bundler (`gem install bundler`)
3. Jekyll (`gem install jekyll`)
4. Once Bundler and Jekyll gems are installed, run `bundle install`

### Run Locally
To construct:
1. `git clone https://github.com/conjurinc/secretless.git`
2. `cd docs`
3. Run the following command:
`bundle exec jekyll serve`
4. Preview Jekyll site locally in web browser by either running `open localhost:4000` or manually navigating to http://localhost:4000