/*
 * ScoreBoard
 * Copyright (C) 2012-2013 Frank Bille
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package dk.frankbille.scoreboard;

import org.eclipse.jetty.proxy.ProxyServlet;
import org.eclipse.jetty.server.Server;
import org.eclipse.jetty.server.handler.ResourceHandler;
import org.eclipse.jetty.servlet.ServletContextHandler;
import org.eclipse.jetty.servlet.ServletHolder;

/**
 * Acts as both a static file server, as well as a proxy for the appengine server.
 */
public class StaticFileServer {

    public static void main(String[] args) throws Exception {
        Server server = new Server(8081);

        ServletContextHandler handler = new ServletContextHandler(server, "/");
        handler.setContextPath("/");
        handler.setResourceBase("src/main/webapp");
        ResourceHandler resourceHandler = new ResourceHandler();
        resourceHandler.setCacheControl("no-cache");
        handler.setHandler(resourceHandler);
        server.setHandler(handler);

        ServletHolder servlet = new ServletHolder(ProxyServlet.Transparent.class);
        server.setAttribute("proxyTo", "http://localhost:8080/api");
        server.setAttribute("prefix", "/api");
        handler.addServlet(servlet, "/api/*");

        server.start();
        server.join();
    }

}
