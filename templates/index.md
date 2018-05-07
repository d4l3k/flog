# Flog

An automated golf registration system.

## Make Reservation

This will attempt to make a reservation at the earliest
possible time (typically 8am).

<form method="post" action="/reserve">
  <table>
    <tbody>
      <tr>
        <td>
          <label for="date">Day</label>
        </td>
        <td>
          <input type="datetime-local" id="date" name="date" value="{{.DefaultDay}}">
        </td>
      </tr>
      <tr>
        <td>
          <label for="players">Number of Players</label>
        </td>
        <td>
          <input type="number" id="players" name="players" value="2" min=1 max=4>
        </td>
      </tr>
      <tr>
        <td></td>
        <td>
          <button type="submit">Schedule Reservation</button>
        </td>
      </tr>
    </tbody>
  </table>
</form>


## Pending Reservations

These are the reservations that will be attempted once they are possible.

<form method="post" action="/cancel">
  <button type="submit">Cancel All Pending Reservations</button>
</form>

{{ range .Pending -}}
* {{.Day}} — {{.Players}} players
{{ else }}
There are no pending reservations.
{{- end }}



## Reservations

You can modify the reservations at: https://www.chronogolf.com/dashboard/#/reservations

{{ range .Reservations -}}
* {{.Teetime.Date}} {{.Teetime.StartTime}} — {{.State}} — {{len .Rounds}} players
{{ else }}
There are no reservations found.
{{- end }}
