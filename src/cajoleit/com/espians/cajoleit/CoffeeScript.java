// No Copyright (-) 2010 The Ampify Authors. This file is under the
// Public Domain license that can be found in the root LICENSE file.

package com.espians.cajoleit;

import java.io.IOException;
import java.io.InputStream;
import java.io.InputStreamReader;
import java.io.Reader;
import java.io.UnsupportedEncodingException;

import org.mozilla.javascript.Context;
import org.mozilla.javascript.ContextFactory;
import org.mozilla.javascript.Scriptable;

// -----------------------------------------------------------------------------
// CoffeeScript Compiler
// -----------------------------------------------------------------------------

public class CoffeeScript {

	private static final String coffeeJS = "com/espians/cajoleit/coffee-script.js";
	private static final String compileJS = "CoffeeScript.compile(source, {bare: true});";
	private static final ContextFactory contextFactory = new ContextFactory();
	private Scriptable parent;

	public CoffeeScript() {

		InputStream stream;
		Reader reader;
		Context ctx;

		stream = getClass().getClassLoader().getResourceAsStream(coffeeJS);
		try {
			reader = new InputStreamReader(stream, "UTF-8");
			ctx = contextFactory.enterContext();
			ctx.setOptimizationLevel(-1);
			try {
				parent = ctx.initStandardObjects();
				ctx.evaluateReader(parent, reader, "coffee-script.js", 1, null);
			} catch (IOException e) {
				throw new Error(e);
			} finally {
				Context.exit();
				try {
					reader.close();
				} catch (IOException e) {
					throw new Error(e);
				}
			}
		} catch (UnsupportedEncodingException e) {
			throw new Error(e);
		} finally {
			try {
				stream.close();
			} catch (IOException e) {
				throw new Error(e);
			}
		}

	}

	public String compile (String source) {

		Context ctx;
		Scriptable scope;

		ctx = contextFactory.enterContext();
		try {
			scope = ctx.newObject(parent);
			scope.setParentScope(parent);
			scope.put("source", scope, source);
			return (String)ctx.evaluateString(scope, compileJS, "compilation", 1, null);
		} finally {
			Context.exit();
		}

	}
}
