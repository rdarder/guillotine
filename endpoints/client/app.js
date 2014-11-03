'use strict';

(function() {
	var app = angular.module("gcuts", []);

	app.controller("main", [ '$scope', function(scope) {
		var api = gapi.client.guillotine.guillotine

		scope.refresh = function() {
			api.randomSpec().execute(function(response) {
				scope.setSpec(response.result);
				scope.$apply();
			});
			api.defaultHints().execute(function(response) {
				scope.setHints(response.result);
				scope.$apply();
			});
		}
		scope.run = function() {
			var spec = angular.copy(scope.spec);
			spec.hints = angular.copy(scope.hints);
			api.cut(spec).execute(function(response) {
				scope.results = response.result;
				scope.$apply();
			});
		}
		scope.setSpec = function(spec) {
			spec.limitedWidth = (spec.maxWidth > 0);
			scope.spec = spec;
			scope.results_old = true;
		}
		scope.setHints = function(hints) {
			scope.hints = hints;
		}
		
		scope.refresh();
	} ]);

	app.directive("resultsPage", function($window) {
		return {
			templateUrl : "resultsPage.html",
			scope : {
				results : '=',
			},
			link: function(scope, element, attrs){
				var toAppend = angular.element("<canvas></canvas>")
				element.append(toAppend);
				var canvas = toAppend[0];
				var ctx = canvas.getContext('2d');
				var ratio
				//take a mostly square letter, take the width and assume that's the height 
				//for the common numbered labels we're using.
				var lineHeight = ctx.measureText('W').width * 1.2;
				var resize = function(){
					ctx.clearRect(0,0, canvas.width, canvas.height);
					if (!scope.results){
						return;
					}
					var sheet = scope.results.sheet;
					var wratio = element[0].offsetWidth / sheet.width;
					var hratio = $window.innerHeight / sheet.height;
					ratio = Math.min(wratio, hratio) * 0.95;

					canvas.width = sheet.width * ratio;
					canvas.height = sheet.height * ratio;
					redraw();
				};
				angular.element($window).bind('resize', function(){
					resize();
				});
				scope.$watch('results', function(){
					resize();
				});
				var randomColor = function(){
				    return '#'+Math.floor(Math.random()*16777215).toString(16);
				};
				var redraw = function(){
					ctx.font = "20pt Arial";
					ctx.fillStyle = "grey";
					ctx.fillRect(0,0, canvas.width, canvas.height);
					angular.forEach(scope.results.boardPlacements, function(placement, i){
						var width = placement.board.width * ratio;
						var height = placement.board.height * ratio;
						var x = placement.placement.x * ratio;
						var y = placement.placement.y * ratio;
						ctx.fillStyle = randomColor();
						ctx.fillRect(x,y,width,height);
						ctx.fillStyle = 'black';
						var label = (i+1).toString();
						var tm = ctx.measureText(label);
						ctx.fillText(label, x+(width-tm.width)/2, y+(height+lineHeight)/2); 
					});
				};
			}
		}
	});
	app.directive("configEditor", function() {
		return {
			templateUrl : "configEditor.html",
			scope : {
				spec : '=',
				hints : '=',
			},
			controller : function($scope) {
				$scope.tab = 'spec';
			},
			transclude : true
		}
	});
	app.directive("hintsEditor", function() {
		return {
			templateUrl : "hintsEditor.html",
			scope : {
				hints : '=',
			}
		}
	});
	app.directive("specEditor", function() {
		return {
			templateUrl : "specEditor.html",
			scope : {
				spec : '=',
			}
		}
	});

	app.directive("boardOrders", function() {
		return {
			templateUrl : "boardOrders.html",
			scope : {
				orders : '='
			},
			controller : function($scope) {
				$scope.remove = function(i) {
					$scope.orders.splice(i, 1);
				}
				$scope.add = function() {
					$scope.orders.push({
						amount : 1,
						width : 0,
						height : 0
					});
				}
			}
		}
	});
	app.directive("boardOrder", function() {
		return {
			templateUrl : "boardOrder.html",
			scope : {
				order : '='
			},
			transclude : true
		}
	})
}());