package sses

import (
	"bufio"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
)

// SSEEvent is a struct that represents an SSE event
type FiberSSEEvent struct {
	Timestamp time.Time `json:"timestamp"`
	ID        string    `json:"id"`
	Event     string    `json:"event"`
	Data      string    `json:"data"`
	Retry     string    `json:"retry"`
	OnChannel *FiberSSEChannel
}

// # TypeDef
//
// Exec for channel events
type FiberSSEEventHandler func(ctx *fiber.Ctx, sseChannel *FiberSSEChannel)

// # TypeDef
//
// Exec for specific events on a channel
type FiberSSEOnEventHandler func(ctx *fiber.Ctx, sseChannel *FiberSSEChannel, sseEvent *FiberSSEEvent)
type FiberSSEEvents interface {
	OnConnect(handlers ...FiberSSEEventHandler)
	OnDisconnect(handlers ...FiberSSEEventHandler)
	OnEvent(eventName string, handlers ...FiberSSEOnEventHandler)
	FireOnEventHandlers(fiberCtx *fiber.Ctx, event string)
}

/*
A channel with a name, and a sub-base-path
*/
type FiberSSEChannel struct {
	FiberSSEEvents
	Name          string
	Base          string
	Events        chan *FiberSSEEvent
	ParentSSEApp  *FiberSSEApp
	Handlers      map[string]([]FiberSSEEventHandler)
	EventHandlers map[string]([]FiberSSEOnEventHandler)
}
type FiberSSEHandler func(c *fiber.Ctx, w *bufio.Writer) error

/*
The SSE Information Structure includes a list of channels and the fiber application
*/
type FiberSSEApp struct {
	IFiberSSEApp
	Base     string
	Router   *fiber.Router
	Channels map[string]*FiberSSEChannel
	FiberApp *fiber.App
}

// FiberSSEApp Interface
type IFiberSSEApp interface {
	ServeHTTP(ctx *fiber.Ctx) error
	CreateChannel(name, base string) *FiberSSEChannel
	ListChannels() map[string]*FiberSSEChannel
	GetChannel(name string) *FiberSSEChannel
}

/*
New initializes a base SSE route group at `base`.

The base route is the base path for all channels.

The channels parameter is a list of channels that will be created.
Each channel has a name, a base route, and a channel for sending events.

	// Create a new SSE app
	app := fiber.New()
	// Create a new SSE app on the fiber app
	sseApp := ssefiber.New(app, "/sse")
	// Add a channel to the SSE app
	testChan := sseApp.CreateChannel("test", "/test") // Channel at /sse/test
	// Events Channel
	eventsChan := testChan.Events
*/
func New(app *fiber.App, base string) *FiberSSEApp {
	// Add the base route
	fiberRouter := app.Group(base, func(c *fiber.Ctx) error {
		// Set the headers for SSE
		c.Set("Cache-Control", "no-cache")
		c.Set("Content-Type", "text/event-stream")
		c.Set("Connection", "keep-alive")
		c.Set("Access-Control-Allow-Origin", "*")
		return c.Next()
	})

	// Create a new SSE App
	newFSSEApp := &FiberSSEApp{
		Base:     base,
		Router:   &fiberRouter,
		FiberApp: app,
		Channels: make(map[string]*FiberSSEChannel),
	}
	return newFSSEApp
}

/*
CreateChannel creates a new channel with the given name and base path.
Functions as a shortcut for making a new chan each time

Example:

	app := fiber.New()
	sseApp := ssefiber.New(app, "/sse")
	chanOne := sseApp.CreateChannel("Channel One", "/one")
	chanTwo := sseApp.CreateChannel("Channel Two", "/two")
*/
func (app *FiberSSEApp) CreateChannel(name, base string) *FiberSSEChannel {
	newChannel := &FiberSSEChannel{
		Name:          name,
		Base:          base,
		Events:        make(chan *FiberSSEEvent),
		ParentSSEApp:  app,
		Handlers:      make(map[string][]FiberSSEEventHandler),
		EventHandlers: make(map[string][]FiberSSEOnEventHandler),
	}
	app.Channels[name] = newChannel
	// Add the sub-route for the channel
	(*app.Router).Get(newChannel.Base, newChannel.ServeHTTP)
	return newChannel
}

// ListChannels returns a list of all the channels and prints them to the console
func (app *FiberSSEApp) ListChannels() map[string]*FiberSSEChannel {
	fmt.Println("Listing Channels...")
	for _, channel := range app.Channels {
		channel.Print()
	}
	return app.Channels
}

/*
Create an event and send it to the channel.
*/
func (channel *FiberSSEChannel) SendEvent(event, data string) {
	sseEvent := &FiberSSEEvent{
		Timestamp: time.Now(),
		Event:     event,
		Data:      data,
		OnChannel: channel,
	}
	channel.Events <- sseEvent
}

// Flush the event to the writer `w` - formats according to SSE standard
func (e *FiberSSEEvent) Flush(w *bufio.Writer) error {
	fmt.Fprintf(w, "event: %s\ndata: %s\n\n", e.Event, e.Data)
	return w.Flush()
}

// Prints the channel information to the console
func (c *FiberSSEChannel) Print() {
	fmt.Printf("==CHANNEL CREATED==\nName: %s\nRoute Endpoint: %s\n===================", c.Name, c.ParentSSEApp.Base+c.Base)
}

// # Internal Method
//
// ServeHTTP returns a fiber.Handler for the channel.
//
// Use `sseApp.CreateChannel` to create a new channel.
func (fChan *FiberSSEChannel) ServeHTTP(c *fiber.Ctx) error {

	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		// Fire OnConnect Event Handlers

		go fChan.FireHandlers(c, "connect")

		for {
			event, more := <-fChan.Events
			// fmt.Fprintf(w, "event: %s\ndata: %s\n\n", string(event.Event), string(event.Data))
			// w.Flush()
			go event.FireEventHandlers(c)
			if err := event.Flush(w); err != nil {
				go fChan.FireHandlers(c, "disconnect")
				return
			}
			if !more {
				// Fire OnDisconnect Event Handlers
				go fChan.FireHandlers(c, "disconnect")
				return
			}
		}
	})

	return nil

}

// Cleanup removes all of the channels from the app. Should be used as a defer
func (sseApp *FiberSSEApp) Cleanup() {
	for _, channel := range sseApp.Channels {
		close(channel.Events)
	}
	fmt.Println("All Channels Closed - Cleanup Successful")
}

// Fire the handlers for a given channel event (connect, disconnect)
func (channel *FiberSSEChannel) FireHandlers(fiberCtx *fiber.Ctx, event string) {
	for _, handler := range channel.Handlers[event] {
		handler(fiberCtx, channel)
	}
}

// Fire the handlers for this event
func (e *FiberSSEEvent) FireEventHandlers(fiberCtx *fiber.Ctx) {
	channel := e.OnChannel
	for _, handler := range channel.EventHandlers[e.Event] {
		handler(fiberCtx, channel, e)
	}
}

// Adds the handlers to the channel for the connect method
func (channel *FiberSSEChannel) OnConnect(handlers ...FiberSSEEventHandler) {
	channel.Handlers["connect"] = []FiberSSEEventHandler{}
	channel.Handlers["connect"] = append(channel.Handlers["connect"], handlers...)
}

// Adds the handlers to the channel for the disconnect method
func (channel *FiberSSEChannel) OnDisconnect(handlers ...FiberSSEEventHandler) {
	channel.Handlers["disconnect"] = []FiberSSEEventHandler{}
	channel.Handlers["disconnect"] = append(channel.Handlers["disconnect"], handlers...)
}

// Add handlers for the any given event
//
// Example:
//
//	channelOne.OnEvent("test", ...) // Fires anytime the event "test" is fired
func (channel *FiberSSEChannel) OnEvent(eventName string, handlers ...FiberSSEOnEventHandler) {
	channel.EventHandlers[eventName] = []FiberSSEOnEventHandler{}
	channel.EventHandlers[eventName] = append(channel.EventHandlers[eventName], handlers...)

}

// Returns a channel by name
func (app *FiberSSEApp) GetChannel(name string) *FiberSSEChannel {
	findChan := app.Channels[name]
	return findChan
}
