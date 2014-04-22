
/**
 * Module dependencies.
 */

var express = require('express')
  , http = require('http')
  , path = require('path');

var app = express();

// all environments
app.set('port', process.env.PORT || 80);
app.set('views', __dirname + '/views');
app.set('view engine', 'ejs');

app.use(express.favicon(path.join(__dirname, 'public/img/favicon.ico'))); 
app.use(express.logger('dev'));
app.use(express.bodyParser());
app.use(express.methodOverride());
app.use(app.router);
app.use(express.static(path.join(__dirname, 'public')));

// development only
if ('development' == app.get('env')) {
  app.use(express.errorHandler());
}

app.get('/', function(req, res){
    res.render('index');
});
app.get('/:docId', function(req, res){
    var centralHostPort = process.argv[2] ? process.argv[2] : "localhost:3000";
    res.render('doc', { docId: req.params.docId, centralHostPort: centralHostPort})
})


http.createServer(app).listen(app.get('port'), function(){
  console.log('Express server listening on port ' + app.get('port'));
});
