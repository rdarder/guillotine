'use strict';

(function() {
	var app = angular.module("gcuts", []);

	app.controller("main", [ '$scope', '$q', '$http', function(scope, $q, $http) {
		var api = gapi.client.guillotine.guillotine

		scope.reset = function() {
			scope.clearApiErrors();
			return $q.all([ scope.getRandomSpec(), scope.getDefaultHints() ]);
		}
		
		scope.onApiError = function(error){
			scope.apiErrors.push(error);
		}
		scope.clearApiErrors = function(){
			scope.apiErrors = [];
		}

		scope.getRandomSpec = function() {
			var pspec = $q.defer();
			api.randomSpec().execute(function(response) {
				if (response.error){
					pspec.reject(scope.onApiError(response.error));
				}
				scope.setSpec(response.result);
				pspec.resolve();
				scope.$apply();
			}, scope.onApiError);
			return pspec.promise;
		}

		scope.getDefaultHints = function() {
			var phints = $q.defer();
			api.defaultHints().execute(function(response) {
				if (response.error){
					phints.reject(scope.onApiError(response.error));
				}
				scope.setHints(response.result);
				phints.resolve();
				scope.$apply();
			}, scope.onApiError);
			return phints.promise;
		}
		scope.run = function() {
			var prun = $q.defer();
			api.cut(scope.getFullSpec()).execute(function(response) {
				if (response.error){
					prun.reject(scope.onApiError(response.error));
				}
				scope.setResults(response.result);
				prun.resolve();
				scope.$apply();
			});
			return prun.promise;
		}
		scope.random = function() {
			scope.getRandomSpec().then(scope.run);
		}
		scope.resetHints = function() {
			scope.getDefaultHints().then(scope.run);
		}
		scope.getFullSpec = function() {
			var spec = angular.copy(scope.spec)
			if (!spec.limitedWidth) {
				spec.maxWidth = 0;
			}
			delete spec.limitedWidth;
			spec.hints = angular.copy(scope.hints);
			return spec
		}
		scope.setSpec = function(spec) {
			spec.limitedWidth = (spec.maxWidth > 0);
			scope.spec = spec;
			scope.results_old = true;
		}
		scope.setHints = function(hints) {
			scope.hints = hints;
		}
		scope.setResults = function(results){
			scope.results = results;
			scope.results_old = false;
		}
		scope.loadSample = function(){
			var url = '/sample' + Math.ceil(Math.random()*4) + '.json';
			$http.get(url).then(function(response){
				scope.setHints(response.data.hints);
				scope.setSpec(response.data.spec);
				scope.setResults(response.data.results);
			});
		}
		
		//load an already processed sample on startup.
		scope.clearApiErrors();
		scope.loadSample();
	} ]);

	app.directive('resultsPage', function(){
		return {
			templateUrl: 'resultsPage.html',
			scope: {
				results: '=',
			}
		}
	})
	app.directive("layoutDisplay", function($window) {
		return {
			link : function(scope, element, attrs) {
				var toAppend = angular.element("<canvas></canvas>")
				element.append(toAppend);
				var canvas = toAppend[0];
				var ctx = canvas.getContext('2d');
				var ratio
				var bWidth = 3;
				var resize = function() {
					ctx.clearRect(0, 0, canvas.width, canvas.height);
					if (!scope.results) {
						return;
					}
					var sheet = scope.results.sheet;
					var wratio = (element[0].offsetWidth - 2 * bWidth)
							/ sheet.width;
					var hratio = ($window.innerHeight - 2 * bWidth)
							/ sheet.height;
					ratio = Math.min(wratio, hratio);

					canvas.width = sheet.width * ratio + 2 * bWidth;
					canvas.height = sheet.height * ratio + 2 * bWidth;
					+100;
					redraw();
				};
				angular.element($window).bind('resize', function() {
					resize();
				});
				scope.$watch('results', function() {
					resize();
				});
				var randomColor = function() {
					var rgb = hslToRgb(Math.random(), 0.45, 0.60);
					return "rgb(" + rgb.join(',') + ')'
				};
				var redraw = function() {
					ctx.save()
					ctx.translate(bWidth, bWidth);
					ctx.strokeStyle = "black"
					ctx.lineWidth = bWidth
					ctx.strokeRect(0, 0, scope.results.sheet.width * ratio,
							scope.results.sheet.height * ratio);
					var labels = []
					angular.forEach(scope.results.boardPlacements, function(p,
							i) {
						rect(p.placement.x * ratio, p.placement.y * ratio,
								p.board.width * ratio, p.board.height * ratio);
						var label = formatLabel(p.board.width, p.board.height,
								p.board.rotated);
						var labelOp = sizeLabel(p.placement.x * ratio,
								p.placement.y * ratio, p.board.width * ratio,
								p.board.height * ratio, label);
						if (labelOp !== null) {
							labels.push(labelOp);
						}
					});
					ctx.fillStyle = "black";
					angular.forEach(labels, drawLabel);
					ctx.restore();

				};
				var drawLabel = function(label) {
					ctx.save();
					ctx.font = label.fontSize + "pt Arial";
					ctx.font
					if (label.orientation === 'v') {
						ctx.translate(label.x, label.y);
						ctx.rotate(-Math.PI / 2);
						ctx.fillText(label.label, 0, 0);
					} else {
						ctx.fillText(label.label, label.x, label.y);
					}
					ctx.restore();
				};
				var rect = function(x, y, width, height, ratio) {
					ctx.fillStyle = randomColor();
					ctx.fillRect(x, y, width, height);
					ctx.strokeRect(x, y, width, height);
				}
				var formatLabel = function(width, height, rot) {
					return [ width, 'x', height, rot ].join(' ');
				}
				var sizeLabel = function(x, y, width, height, label) {
					ctx.save();
					for (var size = 20; size >= 8; size -= 2) {
						ctx.font = size + "pt Arial"
						var textWidth = ctx.measureText(label).width;
						// no height metric available.
						var textHeight = ctx.measureText("0").width; 
						if (textWidth < width - 2 * bWidth
								&& textHeight < height - 2 * bWidth) {
							ctx.restore()
							return {
								label : label,
								x : x + (width - textWidth) / 2 ,
								y : y + (height + textHeight) / 2 ,
								fontSize : size,
								orientation : 'h',
							};
						} else if (textWidth < height - 2 * bWidth
								&& textHeight < width - 2 * bWidth) {
							ctx.restore()
							return {
								label : label,
								x : x + (width + textHeight) / 2 ,
								y : y + (height + textWidth) / 2 ,
								fontSize : size,
								orientation : 'v',
							};
						}
					}
					ctx.restore()
					return null;
				}
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

	// misc utils
	/**
	 * taken from stackoverflow:
	 * http://stackoverflow.com/questions/2353211/hsl-to-rgb-color-conversion
	 * 
	 * 
	 * Converts an HSL color value to RGB. Conversion formula adapted from
	 * http://en.wikipedia.org/wiki/HSL_color_space. Assumes h, s, and l are
	 * contained in the set [0, 1] and returns r, g, and b in the set [0, 255].
	 * 
	 * @param Number
	 *            h The hue
	 * @param Number
	 *            s The saturation
	 * @param Number
	 *            l The lightness
	 * @return Array The RGB representation
	 */
	function hslToRgb(h, s, l) {
		function hue2rgb(p, q, t) {
			if (t < 0)
				t += 1;
			if (t > 1)
				t -= 1;
			if (t < 1 / 6)
				return p + (q - p) * 6 * t;
			if (t < 1 / 2)
				return q;
			if (t < 2 / 3)
				return p + (q - p) * (2 / 3 - t) * 6;
			return p;
		}
		var r, g, b;

		if (s == 0) {
			r = g = b = l; // achromatic
		} else {
			var q = l < 0.5 ? l * (1 + s) : l + s - l * s;
			var p = 2 * l - q;
			r = hue2rgb(p, q, h + 1 / 3);
			g = hue2rgb(p, q, h);
			b = hue2rgb(p, q, h - 1 / 3);
		}

		return [ Math.round(r * 255), Math.round(g * 255), Math.round(b * 255) ];
	}

}());